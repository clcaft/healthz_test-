package dto

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ExampleRequest struct {
	Value int64 `json:"value" validate:"required,GT=0"`
}

type ExampleResponse struct {
	NewValue int64 `json:"newValue" validate:"required,GT=0"`
}
