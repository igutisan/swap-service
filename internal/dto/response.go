package dto

type APIResponse struct {
	Success bool        `json:"success"`
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type ResponseOption func(*APIResponse)

func WithData(data interface{}) ResponseOption {
	return func(r *APIResponse) {
		r.Data = data
	}
}

func WithMessage(msg string) ResponseOption {
	return func(r *APIResponse) {
		r.Message = msg
	}
}

func WithStatus(status int) ResponseOption {
	return func(r *APIResponse) {
		r.Status = status
	}
}

func WithErrors(errors interface{}) ResponseOption {
	return func(r *APIResponse) {
		r.Success = false
		r.Errors = errors
	}
}

func SuccessResponse(opts ...ResponseOption) APIResponse {
	resp := APIResponse{
		Success: true,
		Status:  200,
	}
	for _, opt := range opts {
		opt(&resp)
	}
	return resp
}

func ErrorResponse(status int, message string, opts ...ResponseOption) APIResponse {
	resp := APIResponse{
		Success: false,
		Status:  status,
		Message: message,
	}
	for _, opt := range opts {
		opt(&resp)
	}
	return resp
}
