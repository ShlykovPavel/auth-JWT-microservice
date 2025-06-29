package response

import (
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"net/http"
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
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is invalid", v.Field()))
		}

	}
	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}

// RenderResponse sets the HTTP status code and renders the provided body as JSON.
// The status should be a valid HTTP status code (e.g., http.StatusOK).
// The body is any JSON-serializable object, such as resp.Error or a custom DTO.
func RenderResponse(w http.ResponseWriter, r *http.Request, status int, body interface{}) {
	render.Status(r, status)
	if body == nil {
		render.JSON(w, r, struct{}{})
		return
	}
	render.JSON(w, r, body)
}
