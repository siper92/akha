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

type Transition string
type StepName string

type ServiceProvide interface {
	ServiceProvide() ServiceProvide
}

type DataServiceProvide[D any] interface {
	ServiceProvide
	Data() D
	SetData(d D)
	Update(fn func(d D) D)
}

type Step[S ServiceProvide] interface {
	Name() StepName
	Retry() int
	On(t Transition) (StepName, error)
	process(ctx context.Context, d S) (Transition, error)
}

type StepExec[D any] func(ctx context.Context, s DataServiceProvide[D]) (Transition, error)

type Flow[S ServiceProvide] interface {
	Init(steps ...Step[S]) error
	Execute(ctx context.Context, s S) error // run process then move
	//
	current() Step[S]
	process(ctx context.Context, s S) (Transition, error)
	move(transition Transition) error
}
