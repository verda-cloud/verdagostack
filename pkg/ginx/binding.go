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
