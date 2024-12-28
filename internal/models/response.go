package models

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Başarılı response için helper
func SuccessResponse(data interface{}, message string) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// Hata response'u için helper
func ErrorResponse(err string) Response {
	return Response{
		Success: false,
		Error:   err,
	}
}
