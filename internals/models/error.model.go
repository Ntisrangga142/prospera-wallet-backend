package models

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Status  string `json:"status" example:"HTTP Status Error"`
	Code    int    `json:"status_code" example:"400"`
	Error   string `json:"error"  example:"Error Message"`
}

func NewErrorResponse(status, err string, code int) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Status:  status,
		Code:    code,
		Error:   err,
	}
}
