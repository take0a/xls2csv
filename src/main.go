package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/roboninc/xlsx"
	"github.com/roboninc/xlsx/types"
)

// Table は、テーブルの名前、別名、説明が記載された EXCEL の情報
// EXCEL は、ファイル名、シート名で特定し、１テーブル１行であること。Name が無い場合は無視する
// Name、Alias、Description には、記載された列名（A, B, C...）を設定する
type Table struct {
	Book        string   `json:"book"`
	Sheet       string   `json:"sheet"`
	Name        string   `json:"name"`
	Alias       string   `json:"alias"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// Column は、カラムの名前、別名、説明が記載された EXCEL の情報
// EXCEL は、ファイル名、シート名で特定し、１カラム１行であること。Table が無い場合は無視する
// Table、Name、Alias、Description には、記載された列名（A, B, C...）を設定する
type Column struct {
	Book        string   `json:"book"`
	Sheet       string   `json:"sheet"`
	Table       string   `json:"table"`
	Name        string   `json:"name"`
	Alias       string   `json:"alias"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// Settings は、このプログラムの設定情報
// Input は、Mashu からエクスポートされた CSV ファイル名
// Output は、このプログラムが出力し、Mashu へインポートする CSV ファイル名
type Settings struct {
	Tag     bool     `json:"tag"`
	Input   string   `json:"input"`
	Output  string   `json:"output"`
	Tables  []Table  `json:"tables"`
	Columns []Column `json:"columns"`
}

// Info は、キーに紐づいた値
// テーブルは、Name がキー
// カラムは、Table.Name がキー
type Info struct {
	Alias       string
	Description string
}

// main は、メイン。引数は一つ。JSON 形式の設定ファイルのパス
func main() {
	if len(os.Args) != 2 {
		fmt.Printf("invalid args: %#v", os.Args)
		os.Exit(1)
	}
	err := run(os.Args[1])
	if err != nil {
		fmt.Printf("Error %v", err)
		os.Exit(2)
	}
}

// run は、主処理。JSON 形式の設定ファイルのパスを引数として実行する
func run(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var settings Settings
	err = json.Unmarshal(b, &settings)
	if err != nil {
		return err
	}

	if settings.Tag {
		return runTag(settings)
	}

	tableMap, err := readTables(&settings)
	if err != nil {
		return err
	}
	columnMap, err := readColumns(&settings)
	if err != nil {
		return err
	}

	input, err := os.Open(settings.Input)
	if err != nil {
		return err
	}
	defer input.Close()
	reader := csv.NewReader(input)
	reader.Comment = '#'
	reader.FieldsPerRecord = -1

	output, err := os.Create(settings.Output)
	if err != nil {
		return err
	}
	defer output.Close()
	writer := csv.NewWriter(output)
	defer writer.Flush()

	var tname string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		record, tname = edit(record, tname, tableMap, columnMap)

		err = writer.Write(record)
		if err != nil {
			return err
		}
	}
	return nil
}

// readTables は、全てのテーブル定義書のテーブルの論理名、説明を読み込む
func readTables(settings *Settings) (map[string]Info, error) {
	tableMap := map[string]Info{}
	for _, table := range settings.Tables {
		err := readTable(table, tableMap)
		if err != nil {
			return nil, err
		}
	}
	return tableMap, nil
}

// readTable は、一つのテーブル定義書のテーブルの論理名、説明を読み込む
func readTable(setting Table, m map[string]Info) error {
	name, _ := types.CellRef(setting.Name + "1").ToIndexes()
	if name < 0 {
		return fmt.Errorf("invalid name column %s", setting.Name)
	}
	alias, _ := types.CellRef(setting.Alias + "1").ToIndexes()
	desc, _ := types.CellRef(setting.Description + "1").ToIndexes()

	excel, err := xlsx.Open(setting.Book)
	if err != nil {
		return err
	}
	defer excel.Close()

	sheet := excel.SheetByName(setting.Sheet)
	if sheet == nil {
		return fmt.Errorf("sheet not found %s.%s", setting.Book, setting.Sheet)
	}

	iter := sheet.Rows()
	for iter.HasNext() {
		_, row := iter.Next()
		info := Info{}
		if alias >= 0 {
			info.Alias = row.Cell(alias).Value()
		}
		if desc >= 0 {
			info.Description = row.Cell(desc).Value()
		}
		key := strings.ToLower(row.Cell(name).Value())
		m[key] = info
	}
	return nil
}

// readColumns は、全てのテーブル定義書のカラムの論理名、説明を読み込む
func readColumns(settings *Settings) (map[string]Info, error) {
	columnMap := map[string]Info{}
	for _, column := range settings.Columns {
		err := readColumn(column, columnMap)
		if err != nil {
			return nil, err
		}
	}
	return columnMap, nil
}

// readColumn は、一つのテーブル定義書のカラムの論理名、説明を読み込む
func readColumn(setting Column, m map[string]Info) error {
	table, _ := types.CellRef(setting.Table + "1").ToIndexes()
	if table < 0 {
		return fmt.Errorf("invalid name column %s", setting.Table)
	}
	name, _ := types.CellRef(setting.Name + "1").ToIndexes()
	if name < 0 {
		return fmt.Errorf("invalid name column %s", setting.Name)
	}
	alias, _ := types.CellRef(setting.Alias + "1").ToIndexes()
	desc, _ := types.CellRef(setting.Description + "1").ToIndexes()

	excel, err := xlsx.Open(setting.Book)
	if err != nil {
		return err
	}
	defer excel.Close()

	sheet := excel.SheetByName(setting.Sheet)
	if sheet == nil {
		return fmt.Errorf("sheet not found %s.%s", setting.Book, setting.Sheet)
	}

	iter := sheet.Rows()
	for iter.HasNext() {
		_, row := iter.Next()
		info := Info{}
		if alias >= 0 {
			info.Alias = row.Cell(alias).Value()
		}
		if desc >= 0 {
			info.Description = row.Cell(desc).Value()
		}
		key := fmt.Sprintf("%s.%s",
			strings.ToLower(row.Cell(table).Value()),
			strings.ToLower(row.Cell(name).Value()))
		m[key] = info
	}
	return nil
}

// edit は、テーブル定義情報を使用して CSV の論理名と説明を補完する
func edit(record []string, tname string, tableMap, columnMap map[string]Info) ([]string, string) {
	if len(record) == 0 {
		return record, tname
	}

	switch record[0] {
	case "20":
		if len(record) < 7 {
			return record, tname
		}
		key := strings.ToLower(record[2])
		table, ok := tableMap[key]
		if !ok {
			return record, key
		}
		record[3] = table.Alias
		record[4] = table.Description
		return record, key
	case "30":
		if len(record) < 8 {
			return record, tname
		}
		key := fmt.Sprintf("%s.%s", tname, strings.ToLower(record[2]))
		column, ok := columnMap[key]
		if !ok {
			return record, tname
		}
		record[3] = column.Alias
		record[4] = column.Description
	}
	return record, tname
}
