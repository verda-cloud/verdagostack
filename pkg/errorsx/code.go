// Copyright 2026 Verda Cloud Oy
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errorsx

import "net/http"

// Predefined standard errors.
var (
	OK                  = &ErrorX{Code: http.StatusOK, Message: ""}
	ErrInternal         = &ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError", Message: "Internal server error."}
	ErrNotFound         = &ErrorX{Code: http.StatusNotFound, Reason: "NotFound", Message: "Resource not found."}
	ErrBind             = &ErrorX{Code: http.StatusBadRequest, Reason: "BindError", Message: "Error occurred while binding the request body to the struct."}
	ErrInvalidArgument  = &ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument", Message: "Argument verification failed."}
	ErrUnauthenticated  = &ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated", Message: "Unauthenticated."}
	ErrPermissionDenied = &ErrorX{Code: http.StatusForbidden, Reason: "PermissionDenied", Message: "Permission denied. Access to the requested resource is forbidden."}
	ErrOperationFailed  = &ErrorX{Code: http.StatusConflict, Reason: "OperationFailed", Message: "The requested operation has failed. Please try again later."}
)
