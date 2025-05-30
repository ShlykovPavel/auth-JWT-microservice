package response

import (
	"fmt"
	"github.com/go-playground/validator"
	"strings"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "ERROR"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, v := range errs {
		switch v.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is required", v.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is invalid", v.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is invalid", v.Field()))
		}

	}
	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}
