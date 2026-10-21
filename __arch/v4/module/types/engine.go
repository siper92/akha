package types

type EngineID string

type ExecEngine interface {
	Init() error
	Load(id EngineID) string
	Exec() error
}
