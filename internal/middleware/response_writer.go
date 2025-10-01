package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

// ResponseWriter wraps gin.ResponseWriter to capture response body
type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// NewResponseWriter creates a new response writer wrapper
func NewResponseWriter(w gin.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
	}
}

// Write captures the response body while writing to the original writer
func (w *ResponseWriter) Write(data []byte) (int, error) {
	// Write to our buffer for caching
	w.body.Write(data)

	// Write to the original response writer
	return w.ResponseWriter.Write(data)
}

// WriteString captures the response body while writing to the original writer
func (w *ResponseWriter) WriteString(s string) (int, error) {
	// Write to our buffer for caching
	w.body.WriteString(s)

	// Write to the original response writer
	return w.ResponseWriter.WriteString(s)
}

// Body returns the captured response body
func (w *ResponseWriter) Body() []byte {
	return w.body.Bytes()
}

// BodyString returns the captured response body as string
func (w *ResponseWriter) BodyString() string {
	return w.body.String()
}

// Reset clears the captured body buffer
func (w *ResponseWriter) Reset() {
	w.body.Reset()
}

// Size returns the size of the captured body
func (w *ResponseWriter) Size() int {
	return w.body.Len()
}

// ResponseCapture middleware captures response body for idempotency caching
func ResponseCapture() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only capture for methods that support idempotency
		method := c.Request.Method
		if method == "POST" || method == "PUT" || method == "PATCH" {
			// Wrap the response writer
			wrapper := NewResponseWriter(c.Writer)
			c.Writer = wrapper

			// Store the wrapper in context for later access
			c.Set("response_writer", wrapper)
		}

		c.Next()
	}
}

// GetResponseWriter safely extracts the response writer wrapper from context
func GetResponseWriter(c *gin.Context) (*ResponseWriter, bool) {
	if wrapper, exists := c.Get("response_writer"); exists {
		if rw, ok := wrapper.(*ResponseWriter); ok {
			return rw, true
		}
	}
	return nil, false
}

// CachedResponseData represents the data to be cached for idempotency
type CachedResponseData struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
	Timestamp  int64               `json:"timestamp"`
}

// NewCachedResponseData creates cached response data from gin context
func NewCachedResponseData(c *gin.Context) *CachedResponseData {
	data := &CachedResponseData{
		StatusCode: c.Writer.Status(),
		Headers:    make(map[string][]string),
		Timestamp:  c.GetInt64("request_timestamp"),
	}

	// Copy headers
	for key, values := range c.Writer.Header() {
		data.Headers[key] = make([]string, len(values))
		copy(data.Headers[key], values)
	}

	// Get body if response writer wrapper is available
	if wrapper, ok := GetResponseWriter(c); ok {
		data.Body = wrapper.BodyString()
	}

	return data
}

// ApplyToContext applies cached response data to gin context
func (d *CachedResponseData) ApplyToContext(c *gin.Context) {
	// Set headers
	for key, values := range d.Headers {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Set status and body
	c.Data(d.StatusCode, c.GetHeader("Content-Type"), []byte(d.Body))
}
