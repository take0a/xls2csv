.PHONY: clean build test

BIN_OUTPUT := $(if $(filter $(shell go env GOOS), windows), dist/xls2csv.exe, dist/xls2csv)

clean:
	rm -f ${BIN_OUTPUT} ./src/testdata/*output.csv

build: clean
	go build -o ${BIN_OUTPUT} ./...

test:
	go test -cover ./...