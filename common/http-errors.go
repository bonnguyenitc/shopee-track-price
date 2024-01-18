package common

import (
	"net/http"
)

type ResponseApi struct {
	Status   int    `json:"status"`
	Message  string `json:"message"`
	Metadata any    `json:"metadata"`
}

type ResponseErrorApi struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func FromError(err error) int {
	switch err {
	case ErrBadRequest:
		return http.StatusBadRequest
	case ErrNotFound:
		return http.StatusNotFound
	case ErrInternalFail:
		return http.StatusInternalServerError
	case ErrForbidden:
		return http.StatusForbidden
	case ErrServiceDown:
		return http.StatusServiceUnavailable
	}
	return http.StatusUnauthorized
}

func ReturnErrorApi(status int, code string, message string) ResponseErrorApi {
	return ResponseErrorApi{
		Status:  status,
		Message: message,
		Code:    code,
	}
}

func ReturnApi(data interface{}, message string) ResponseApi {
	return ResponseApi{
		Status:   http.StatusOK,
		Message:  message,
		Metadata: data,
	}
}
