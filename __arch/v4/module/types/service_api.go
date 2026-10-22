package types

type RestCallTypes string

const (
	APIQueryCallTypeGET    RestCallTypes = "GET"
	APIQueryCallTypePOST   RestCallTypes = "POST"
	APIQueryCallTypePUT    RestCallTypes = "PUT"
	APIQueryCallTypeDELETE RestCallTypes = "DELETE"
)

type ApiQuery interface {
	Endpoint() string
	CallType() RestCallTypes
}

type WorkerAPIService interface {
	Init(opt Config, prov ServiceProvider) error
	Call(query ApiQuery, params ObjVal) (DataUnit, error)
}

type WorkerSpawnerService interface {
	Init(opt Config, prov ServiceProvider) error
	SpawnWorker(id string) (WorkerAPIService, error)
}

type CallRequest interface {
	Call() ApiQuery
	Params() ObjVal
	Validate() error
}

type BackendResponse interface {
	Data() DataUnit
	Error() error
}

type BackendAPIService[A any] interface {
	Init(opt Config, prov ServiceProvider) error
	Answer(call CallRequest) (BackendResponse, error)
	BuildAPI(apiDef FileTypeSource[A]) error
}

type fileName string

type FileTypeSource[T any] interface {
	Source[fileName, T]
}

// rest API definition

type RestDefinitionSource interface {
	Path() string
	GetMethod() RestCallTypes
	Validate(response BackendResponse) error
}

type RestAPIHandler interface {
	Handle(call CallRequest) (BackendResponse, error)
}

type RestAPIDefinition interface {
	GetPort() int
	Find(call CallRequest) (RestAPIHandler, error)
}
