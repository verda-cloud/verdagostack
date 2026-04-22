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

package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// Bind processes data from multiple sources without validating between each step,
// then performs a single validation at the end. Validation is temporarily disabled
// during binding and restored afterward, so other code using Gin's standard
// ShouldBind* methods is not affected.
//
// Example:
//
//	var req UserUpdateRequest
//	if err := ginx.Bind(c, &req, ginx.URI, ginx.JSON); err != nil {
//	    // handle error
//	}
func Bind(c *gin.Context, obj interface{}, bindFuncs ...func(*gin.Context, interface{}) error) error {
	savedValidator := binding.Validator
	binding.Validator = nil
	defer func() { binding.Validator = savedValidator }()

	for _, bindFunc := range bindFuncs {
		if err := bindFunc(c, obj); err != nil {
			return err
		}
	}

	if savedValidator != nil {
		return savedValidator.ValidateStruct(obj)
	}

	return nil
}

// URI binds URI parameters to the given object.
func URI(c *gin.Context, obj interface{}) error {
	return c.ShouldBindUri(obj)
}

// JSON binds JSON body to the given object.
func JSON(c *gin.Context, obj interface{}) error {
	return c.ShouldBindJSON(obj)
}
