package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"huydq2801.com/golearn/internal/utils"
)

type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *CustomResponseWriter) Write(data []byte) (n int, err error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}
func LoggerMiddleware(logger *zerolog.Logger) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		start := time.Now()
		contentType := ctx.GetHeader("Content-Type")
		requestBody := make(map[string]any)
		var formFiles []map[string]any
		if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := ctx.Request.ParseMultipartForm(10 << 20); err == nil && ctx.Request.MultipartForm != nil {
				for k, v := range ctx.Request.MultipartForm.Value {
					if len(v) == 1 {
						requestBody[k] = v[0]
					} else {
						requestBody[k] = v
					}
				}
				for field, files := range ctx.Request.MultipartForm.File {
					for _, file := range files {
						formFiles = append(formFiles, map[string]any{
							"field":        field,
							"file_name":    file.Filename,
							"file_size":    utils.FormatFileSize(file.Size),
							"content_type": file.Header.Get("Content-Type"),
						})
					}
				}
			} else {
				logger.Warn().Err(err).Msg("Failed to parse multipart form")
			}

			if len(formFiles) > 0 {
				requestBody["form_files"] = formFiles
			}
		} else {
			bodyBytes, err := io.ReadAll(ctx.Request.Body)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to read request body")
			}
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if strings.HasPrefix(contentType, "application/json") {
				_ = json.Unmarshal(bodyBytes, &requestBody)
			} else if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
				value, _ := url.ParseQuery(string(bodyBytes))
				for k, v := range value {
					if len(v) == 1 {
						requestBody[k] = v[0]
					} else {
						requestBody[k] = v
					}
				}
			} else if strings.HasPrefix(contentType, "text/plain") {
				requestBody["data"] = string(bodyBytes)
			} else {
				requestBody["data"] = string(bodyBytes)
			}
		}
		customWriter := &CustomResponseWriter{
			ResponseWriter: ctx.Writer,
			body:           bytes.NewBufferString(""),
		}
		ctx.Writer = customWriter

		ctx.Next()

		duration := time.Since(start)

		statusCode := ctx.Writer.Status()

		responseContentType := ctx.Writer.Header().Get("Content-Type")
		responseBodyRaw := customWriter.body.String()
		var responseBodyParsed interface{}
		if strings.HasPrefix(responseContentType, "image/") {
			responseBodyParsed = "Image Data"
		} else if strings.HasPrefix(responseContentType, "application/json") ||
			strings.HasPrefix(strings.TrimSpace(responseBodyRaw), "{") ||
			strings.HasPrefix(strings.TrimSpace(responseBodyRaw), "[") {
			if err := json.Unmarshal([]byte(responseBodyRaw), &responseBodyParsed); err != nil {
				responseBodyParsed = responseBodyRaw
			}
		} else {
			responseBodyParsed = responseBodyRaw
		}

		logEvent := logger.Info()
		if statusCode >= 500 {
			logEvent = logger.Error()
		} else if statusCode >= 400 {
			logEvent = logger.Warn()
		}
		logEvent.
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Str("query", ctx.Request.URL.RawQuery).
			Str("client_ip", ctx.ClientIP()).
			Str("user_agent", ctx.Request.UserAgent()).
			Str("referer", ctx.Request.Referer()).
			Str("protocol", ctx.Request.Proto).
			Str("host", ctx.Request.Host).
			Str("remote_addr", ctx.Request.RemoteAddr).
			Str("request_uri", ctx.Request.RequestURI).
			Int64("content_length", ctx.Request.ContentLength).
			Interface("headers", ctx.Request.Header).
			Interface("request_body", requestBody).
			Interface("response_body", responseBodyParsed).
			Int64("status_code", int64(ctx.Writer.Status())).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("HTTP Request Log")
	}

}
