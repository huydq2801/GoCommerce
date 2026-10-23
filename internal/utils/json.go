package utils

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// DecodeJSON rejects unknown fields, multiple JSON values, and oversized bodies.
func DecodeJSON(c *gin.Context, dst any) bool {
	mediaType, _, mediaErr := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if mediaErr != nil || mediaType != "application/json" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": gin.H{"code": "UNSUPPORTED_MEDIA_TYPE", "message": "Use application/json"}})
		return false
	}
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(dst)
	if err == nil {
		var extra any
		if e := decoder.Decode(&extra); e != io.EOF {
			if e != nil {
				err = e
			} else {
				err = errors.New("expected one JSON value")
			}
		}
	}
	if err == nil {
		err = binding.Validator.ValidateStruct(dst)
	}
	if err != nil {
		var sizeErr *http.MaxBytesError
		if errors.As(err, &sizeErr) {
			c.JSON(413, gin.H{"error": gin.H{"code": "PAYLOAD_TOO_LARGE", "message": "Request body exceeds 1 MiB"}})
		} else {
			ResponseError(c, NewError("Invalid JSON or field values", ErrCodeBadRequest))
		}
		return false
	}
	return true
}
