package types

type NotFoundResponse struct {
	Error string `json:"error"`
}

type BadRequestResponse struct {
	Error string `json:"error"`
}

type BadRequestStructResponse struct {
	Error []string `json:"error"`
}

type ContextCanceledResponse struct {
	Error string `json:"error"`
}

type InternalServerErrorResponse struct {
	Error string `json:"error"`
}

type UnauthorizedResponse struct {
	Error string `json:"error"`
}

type ErrorResponse interface {
	NotFoundResponse |
		BadRequestResponse |
		ContextCanceledResponse |
		InternalServerErrorResponse |
		BadRequestStructResponse |
		UnauthorizedResponse
}
