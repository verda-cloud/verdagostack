// Package ginx provides Gin-specific extensions: multi-source request binding,
// typed request handlers with validation, and structured error responses.
package ginx

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/verda-cloud/verdagostack/pkg/errorsx"
)

// Validator is a validation function applied after request binding.
type Validator[T any] func(context.Context, *T) error

// Binder defines a function that binds request data to a struct.
type Binder func(any) error

// Handler is a typed request handler that receives a bound+validated request
// and returns a response or error.
type Handler[T any, R any] func(ctx context.Context, req *T) (R, error)

// ErrorResponse is the standard error envelope returned by WriteResponse.
type ErrorResponse struct {
	Reason   string            `json:"reason,omitempty"`
	Message  string            `json:"message,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// HandleAllRequest binds URI, query/form, and JSON body into a single struct,
// runs validators, then calls the handler.
func HandleAllRequest[T any, R any](c *gin.Context, handler Handler[T, R], validators ...Validator[T]) {
	var request T

	if err := ShouldBindAll(c, &request, validators...); err != nil {
		WriteResponse(c, nil, err)
		return
	}

	response, err := handler(c.Request.Context(), &request)
	WriteResponse(c, response, err)
}

// HandleJSONRequest binds the JSON body, validates, and calls the handler.
func HandleJSONRequest[T any, R any](c *gin.Context, handler Handler[T, R], validators ...Validator[T]) {
	HandleRequest(c, c.ShouldBindJSON, handler, validators...)
}

// HandleQueryRequest binds query parameters, validates, and calls the handler.
func HandleQueryRequest[T any, R any](c *gin.Context, handler Handler[T, R], validators ...Validator[T]) {
	HandleRequest(c, c.ShouldBindQuery, handler, validators...)
}

// HandleUriRequest binds URI parameters, validates, and calls the handler.
func HandleUriRequest[T any, R any](c *gin.Context, handler Handler[T, R], validators ...Validator[T]) {
	HandleRequest(c, c.ShouldBindUri, handler, validators...)
}

// HandleRequest is the generic request handler: bind, validate, call handler,
// then write the response.
func HandleRequest[T any, R any](c *gin.Context, binder Binder, handler Handler[T, R], validators ...Validator[T]) {
	var request T

	if err := ReadRequest(c, &request, binder, validators...); err != nil {
		WriteResponse(c, nil, err)
		return
	}

	response, err := handler(c.Request.Context(), &request)
	WriteResponse(c, response, err)
}

// ShouldBindJSON binds JSON body and runs validators.
func ShouldBindJSON[T any](c *gin.Context, rq *T, validators ...Validator[T]) error {
	return ReadRequest(c, rq, c.ShouldBindJSON, validators...)
}

// ShouldBindQuery binds query parameters and runs validators.
func ShouldBindQuery[T any](c *gin.Context, rq *T, validators ...Validator[T]) error {
	return ReadRequest(c, rq, c.ShouldBindQuery, validators...)
}

// ShouldBindUri binds URI parameters and runs validators.
func ShouldBindUri[T any](c *gin.Context, rq *T, validators ...Validator[T]) error {
	return ReadRequest(c, rq, c.ShouldBindUri, validators...)
}

// ShouldBindAll binds URI, query/form, and JSON body into a single struct,
// then applies Default() and validators.
func ShouldBindAll[T any](c *gin.Context, rq *T, validators ...Validator[T]) error {
	if err := Bind(c, rq, URI, JSON); err != nil {
		return errorsx.ErrBind.WithMessage("%s", err.Error())
	}

	return FinalizeRequest(c, rq, validators...)
}

// ReadRequest binds request data using the provided binder, applies Default(),
// and runs validators.
func ReadRequest[T any](c *gin.Context, rq *T, binder Binder, validators ...Validator[T]) error {
	if err := binder(rq); err != nil {
		return errorsx.ErrBind.WithMessage("%s", err.Error())
	}

	return FinalizeRequest(c, rq, validators...)
}

// FinalizeRequest applies Default() (if the type implements it) and runs all
// validators sequentially.
func FinalizeRequest[T any](c *gin.Context, rq *T, validators ...Validator[T]) error {
	if defaulter, ok := any(rq).(interface{ Default() }); ok {
		defaulter.Default()
	}

	for _, validate := range validators {
		if validate == nil {
			continue
		}
		if err := validate(c.Request.Context(), rq); err != nil {
			return err
		}
	}

	return nil
}

// WriteResponse writes a JSON success (200) or error response.
// Errors are converted to *errorsx.ErrorX to produce a structured error envelope.
func WriteResponse(c *gin.Context, data any, err error) {
	if err != nil {
		errx := errorsx.FromError(err)
		c.JSON(errx.Code, ErrorResponse{
			Reason:   errx.Reason,
			Message:  errx.Message,
			Metadata: errx.Metadata,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}
