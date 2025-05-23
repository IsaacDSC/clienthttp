package structs

type BaseInput struct {
	Endpoint    string
	QueryParams map[string]string
	Headers     map[string]string
}

type GetRequest struct {
	BaseInput
}

type DelRequest struct {
	BaseInput
}

type PatchRequest struct {
	BaseInput
	Body []byte
}

type PostRequest struct {
	BaseInput
	Body []byte
}

type PutRequest struct {
	BaseInput
	Body []byte
}
