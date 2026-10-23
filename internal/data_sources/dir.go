package data_sources

type DirEntityType string

const (
	DirEntityTypeFile DirEntityType = "file"
	DirEntityTypeDir  DirEntityType = "dir"
)

type DirEntityProcessor[T any] interface {
	Process(entity DirEntity) (T, error)
}

type DirEntity interface {
	Name() string
	Type() DirEntityType
	Processor() DirEntityProcessor[any]
}

type DirDataSource interface {
	Tree() (DirEntity, error)
	Find(path string) (DirEntity, error) // find file or dir based on path
}
