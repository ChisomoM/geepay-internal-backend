package response

// Response represents the standard API response structure.
// Every endpoint returns: { "status": 0|1, "message": "...", "data": ... }
type Response struct {
	Status  int         `json:"status"` // 0 = success, 1 = error
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success returns a success response with data.
func Success(data interface{}) Response {
	return Response{
		Status:  0,
		Message: "success",
		Data:    data,
	}
}

// SuccessWithMessage returns a success response with custom message and data.
func SuccessWithMessage(message string, data interface{}) Response {
	return Response{
		Status:  0,
		Message: message,
		Data:    data,
	}
}

// Error returns an error response with null data.
func Error(message string) Response {
	return Response{
		Status:  1,
		Message: message,
		Data:    nil,
	}
}

// ErrorWithData returns an error response with additional context (e.g., validation errors).
func ErrorWithData(message string, data interface{}) Response {
	return Response{
		Status:  1,
		Message: message,
		Data:    data,
	}
}
