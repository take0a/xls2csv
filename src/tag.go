package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/roboninc/xlsx"
	"github.com/roboninc/xlsx/types"
)

// runTag は、主処理。JSON 形式の設定ファイルのパスを引数として実行する
func runTag(settings Settings) error {
	tableMap, err := readTagTables(&settings)
	if err != nil {
		return err
	}
	columnMap, err := readTagColumns(&settings)
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

		record, tname = editTag(record, tname, tableMap, columnMap)

		err = writer.Write(record)
		if err != nil {
			return err
		}
	}
	return nil
}

// readTagTables は、全てのテーブル定義書のテーブルの論理名、説明を読み込む
func readTagTables(settings *Settings) (map[string][]string, error) {
	tableMap := map[string][]string{}
	for _, table := range settings.Tables {
		err := readTagTable(table, tableMap)
		if err != nil {
			return nil, err
		}
	}
	return tableMap, nil
}

// readTagTable は、一つのテーブル定義書のテーブルの論理名、説明を読み込む
func readTagTable(setting Table, m map[string][]string) error {
	name, _ := types.CellRef(setting.Name + "1").ToIndexes()
	if name < 0 {
		return fmt.Errorf("invalid name column %s", setting.Name)
	}

	var tags []int
	for _, col := range setting.Tags {
		tag, _ := types.CellRef(col + "1").ToIndexes()
		if tag >= 0 {
			tags = append(tags, tag)
		}
	}
	if len(tags) == 0 {
		return nil
	}

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
		var strs []string
		for _, col := range tags {
			strs = append(strs, row.Cell(col).Value())
		}
		key := strings.ToLower(row.Cell(name).Value())
		m[key] = strs
	}
	return nil
}

// readTagColumns は、全てのテーブル定義書のカラムの論理名、説明を読み込む
func readTagColumns(settings *Settings) (map[string][]string, error) {
	columnMap := map[string][]string{}
	for _, column := range settings.Columns {
		err := readTagColumn(column, columnMap)
		if err != nil {
			return nil, err
		}
	}
	return columnMap, nil
}

// readTagColumn は、一つのテーブル定義書のカラムの論理名、説明を読み込む
func readTagColumn(setting Column, m map[string][]string) error {
	table, _ := types.CellRef(setting.Table + "1").ToIndexes()
	if table < 0 {
		return fmt.Errorf("invalid name column %s", setting.Table)
	}
	name, _ := types.CellRef(setting.Name + "1").ToIndexes()
	if name < 0 {
		return fmt.Errorf("invalid name column %s", setting.Name)
	}

	var tags []int
	for _, col := range setting.Tags {
		tag, _ := types.CellRef(col + "1").ToIndexes()
		if tag >= 0 {
			tags = append(tags, tag)
		}
	}
	if len(tags) == 0 {
		return nil
	}

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
		var strs []string
		for _, col := range tags {
			strs = append(strs, row.Cell(col).Value())
		}
		key := fmt.Sprintf("%s.%s",
			strings.ToLower(row.Cell(table).Value()),
			strings.ToLower(row.Cell(name).Value()))
		m[key] = strs
	}
	return nil
}

// editTag は、テーブル定義情報を使用して CSV の論理名と説明を補完する
func editTag(record []string, tname string, tableMap, columnMap map[string][]string) ([]string, string) {
	if len(record) == 0 {
		return record, tname
	}

	switch record[0] {
	case "41":
		key := strings.ToLower(record[2])
		tags, ok := tableMap[key]
		if !ok {
			return record, key
		}
		if len(record) != len(tags)+4 {
			return record, key
		}
		for i, str := range tags {
			record[i+4] = str
		}
		return record, key
	case "42":
		key := fmt.Sprintf("%s.%s", tname, strings.ToLower(record[2]))
		tags, ok := columnMap[key]
		if !ok {
			return record, tname
		}
		if len(record) != len(tags)+4 {
			return record, key
		}
		for i, str := range tags {
			record[i+4] = str
		}
		return record, tname
	}
	return record, tname
}
