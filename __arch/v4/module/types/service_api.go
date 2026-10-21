package types

type APIQueryCallType string

const (
	APIQueryCallTypeGET    APIQueryCallType = "GET"
	APIQueryCallTypePOST   APIQueryCallType = "POST"
	APIQueryCallTypePUT    APIQueryCallType = "PUT"
	APIQueryCallTypeDELETE APIQueryCallType = "DELETE"
)

type ApiQuery interface {
	Endpoint() string
	CallType() APIQueryCallType
}

type FeAPIService interface {
	Init(opt Config, prov ServiceProvider) error
	Call(query ApiQuery, params ObjVal) (DataUnit, error)
}
