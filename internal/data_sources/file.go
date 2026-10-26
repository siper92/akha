package data_sources

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrExtension = errors.New("unexpected extension")

type fileBase struct {
	path string
}

func (f fileBase) Path() string { return f.path }

func (f fileBase) Ext() string { return strings.TrimPrefix(filepath.Ext(f.path), ".") }

func (f fileBase) expect(ext, want string) error {
	if ext != want {
		return fmt.Errorf("%w: %s want %s", ErrExtension, ext, want)
	}
	return nil
}

var _ FileDataSource[string] = (*TextFile)(nil)

type TextFile struct {
	fileBase
}

func NewTextFile(path string) File { return &TextFile{fileBase{path}} }

func (f *TextFile) Init(ext string) error { return f.expect(ext, "txt") }

func (f *TextFile) Process(context.Context) (string, error) {
	b, err := os.ReadFile(f.path)
	return string(b), err
}

var _ FileDataSource[struct{}] = (*JSONFile[struct{}])(nil)

type JSONFile[T any] struct {
	fileBase
}

func JSONHandler[T any]() FileHandler {
	return func(path string) File { return &JSONFile[T]{fileBase{path}} }
}

func (f *JSONFile[T]) Init(ext string) error { return f.expect(ext, "json") }

func (f *JSONFile[T]) Process(context.Context) (T, error) {
	var v T
	b, err := os.ReadFile(f.path)
	if err != nil {
		return v, err
	}
	err = json.Unmarshal(b, &v)
	return v, err
}

type Row map[string]string

var _ FileDataSource[[]Row] = (*CSVFile)(nil)

type CSVFile struct {
	fileBase
	hasHeader bool
}

func CSVHandler(hasHeader bool) FileHandler {
	return func(path string) File { return &CSVFile{fileBase: fileBase{path}, hasHeader: hasHeader} }
}

func (f *CSVFile) Init(ext string) error { return f.expect(ext, "csv") }

func (f *CSVFile) Process(context.Context) ([]Row, error) {
	r, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	records, err := csv.NewReader(r).ReadAll()
	if err != nil || len(records) == 0 {
		return nil, err
	}
	keys := make([]string, len(records[0]))
	for i := range keys {
		keys[i] = fmt.Sprint(i)
	}
	if f.hasHeader {
		keys, records = records[0], records[1:]
	}
	rows := make([]Row, 0, len(records))
	for _, rec := range records {
		row := make(Row, len(rec))
		for i, v := range rec {
			row[keys[i]] = v
		}
		rows = append(rows, row)
	}
	return rows, nil
}
