package errorsx

import "net/http"

// Predefined standard errors.
var (
	OK                = &ErrorX{Code: http.StatusOK, Message: ""}
	ErrInternal       = &ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError", Message: "Internal server error."}
	ErrNotFound       = &ErrorX{Code: http.StatusNotFound, Reason: "NotFound", Message: "Resource not found."}
	ErrBind           = &ErrorX{Code: http.StatusBadRequest, Reason: "BindError", Message: "Error occurred while binding the request body to the struct."}
	ErrInvalidArgument = &ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument", Message: "Argument verification failed."}
	ErrUnauthenticated = &ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated", Message: "Unauthenticated."}
	ErrPermissionDenied = &ErrorX{Code: http.StatusForbidden, Reason: "PermissionDenied", Message: "Permission denied. Access to the requested resource is forbidden."}
	ErrOperationFailed = &ErrorX{Code: http.StatusConflict, Reason: "OperationFailed", Message: "The requested operation has failed. Please try again later."}
)
