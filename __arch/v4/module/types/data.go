package types

type SttVal Value[string]
type ByteVal Value[[]byte]
type BoolVal Value[bool]
type IntVal Value[int]
type Int64Val Value[int64]
type ObjVal Value[map[string]string]

type Value[T any] interface {
	Get() T
	Set(value T)
}

type DataUnit interface {
	Read() []byte
	Validate() error
}

type Source[ID comparable, T DataUnit] interface {
	Get(id ID) (T, error)
}

type SourcePut[ID comparable, T DataUnit] interface {
	Put(id ID, data T) (bool, error)
}

type SourceUpdate[ID comparable, T DataUnit] interface {
	Update(id ID, data T) (bool, error)
}

type SourceDelete[ID comparable, T DataUnit] interface {
	Delete(id ID) (bool, error)
}

// types of sources

// os
type MimeType string

const (
	MimeTypeTextPlain MimeType = "text/plain"
	MimeTypeTextJSON  MimeType = "application/json"
	MimeTypeTextHTML  MimeType = "text/html"
)

type FileData struct {
	Name     string
	Size     int64
	MimeType MimeType
	Content  []byte
}

type FileSource interface {
	Init(dir string) error
	Source[string, FileData]
	SourcePut[string, FileData]
	SourceUpdate[string, FileData]
	SourceDelete[string, FileData]
}

// db

type Query interface {
	Sql() string
	Params() []string
}

type SqlLiteSource[ID Query, T DataUnit] interface {
	Init(source string) error
	Source[ID, T]
	SourcePut[ID, T]
	SourceUpdate[ID, T]
	SourceDelete[ID, T]
}

// api/http

type RestSource[ID comparable, T DataUnit] interface {
	Init(baseURL string) error
	Source[ID, T]
	SourcePut[ID, T]
	SourceUpdate[ID, T]
	SourceDelete[ID, T]
}
