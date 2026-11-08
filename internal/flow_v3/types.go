package flow_v3

import (
	"context"
	"database/sql"
)

// ####################################
// File System

type File interface {
	Name() string
	Path() string
	Content() []byte
}

type FileSystemAdapter interface {
	Walk(path string, fn func(file File) error) error
	Read(path string) ([]byte, error)
	Move(path string, newPath string) error
}

// ####################################
// SQL

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type DBAdapter interface {
	Conn() (*sql.DB, error)
}

// ####################################
// AI API

type AIAdapter interface {
	Ask(prompt string) (string, error)
}

// ####################################

type SourceProvider interface {
	FileSystemAdapter() FileSystemAdapter
	DBAdapter() DBAdapter
	AIAdapter() AIAdapter
}

type Step interface {
	Name() string
	Retry() int
}

type FlowGraph interface {
	On(step Step) (Step, error)
}

type Flow[T any] interface {
	Graph() FlowGraph
	Data() T
	Execute() error
}
