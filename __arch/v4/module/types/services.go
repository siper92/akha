package types

import "context"

type ServiceID string

type ConfigOption interface {
	DataUnit
	Value() SttVal
}

type Config Source[string, ConfigOption]

type Service interface {
	Init(ctx context.Context, opt Config) error
	ID() ServiceID
	Context() context.Context
}

type ServiceProvider interface {
	Init(ctx context.Context) error
	RegisterService(id ServiceID, service Service) error
	GetService(id ServiceID) (ctx context.Context, service Service, err error)
}

type APIService interface {
	Service
	Api() RestSource[Query, DataUnit]
}

type APICall interface {
	Params() map[string]string
	Param(key string) string
	Type() string
}

type APIServiceFrontend interface {
	Service
	Call(query Query) (DataUnit, error)
}

type DbService interface {
	Service
	Db() SqlLiteSource[Query, DataUnit]
}

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

type LoggerService interface {
	Service
	Msg(message string, args ...interface{}) error
	Info(message string, args ...interface{}) error
	Warn(message string, args ...interface{}) error
	Error(message string, args ...interface{}) error
	log(logLevel LogLevel, message string, args ...interface{}) error
}
