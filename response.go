package doze

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
)

// BodyWriter defines the interface for writing the response body.
type BodyWriter interface {
	WriteBody(w io.Writer) (int, error)
}

// ResponseSender is an interface for sending responses from controller actions.
type ResponseSender interface {
	Send(w io.Writer) (int, error)
}

// PlainBodyWriter writes the body as plain bytes.
type PlainBodyWriter struct {
	Body []byte
}

// WriteBody writes the plain body to the io.Writer.
func (bw *PlainBodyWriter) WriteBody(w io.Writer) (int, error) {
	if len(bw.Body) > 0 {
		return w.Write(bw.Body)
	}
	return 0, nil
}

// BasicResponse contains all basic properties for a response.
type BasicResponse struct {
	StatusCode int
	Headers    http.Header
	BodyWriter BodyWriter
}

// Send writes the response headers and body to the provided io.Writer.
func (r *BasicResponse) Send(w io.Writer) (int, error) {
	// Check if w implements http.ResponseWriter
	if rw, ok := w.(http.ResponseWriter); ok {
		// Set headers
		for key, values := range r.Headers {
			for _, value := range values {
				rw.Header().Set(key, value)
			}
		}
		// Set status code
		if r.StatusCode != 0 {
			rw.WriteHeader(r.StatusCode)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
	}

	// Use BodyWriter if provided
	if r.BodyWriter != nil {
		return r.BodyWriter.WriteBody(w)
	}

	// No body to write
	return 0, nil
}

// NewOKJSONResponse returns a BasicResponse tailored for JSON with status code 200.
func NewOKJSONResponse(body interface{}) *BasicResponse {
	resp := basicJSONResponse(body)
	resp.StatusCode = http.StatusOK
	return resp
}

// NewCreatedJSONResponse returns a BasicResponse tailored for JSON with status code 201.
func NewCreatedJSONResponse(body interface{}) *BasicResponse {
	resp := basicJSONResponse(body)
	resp.StatusCode = http.StatusCreated
	return resp
}

// NewInternalServerErrorResponse returns a BasicResponse with status code 500.
func NewInternalServerErrorResponse() *BasicResponse {
	return &BasicResponse{
		StatusCode: http.StatusInternalServerError,
		BodyWriter: &PlainBodyWriter{Body: []byte(http.StatusText(http.StatusInternalServerError))},
	}
}

// basicJSONResponse creates a BasicResponse with JSON content.
func basicJSONResponse(body interface{}) *BasicResponse {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return &BasicResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
		}
	}

	return &BasicResponse{
		Headers:    headers,
		BodyWriter: &PlainBodyWriter{Body: jsonBody},
	}
}

// GzipResponse wraps a BasicResponse to provide gzip compression.
type GzipResponse struct {
	*BasicResponse
	inner BodyWriter
}

// WriteBody compresses the body using gzip and writes it to the io.Writer.
func (gr *GzipResponse) WriteBody(w io.Writer) (int, error) {
	gz := gzip.NewWriter(w)
	defer gz.Close()
	return gr.inner.WriteBody(gz)
}

// NewGzipResponse wraps a BasicResponse to create a GzipResponse.
func NewGzipResponse(br *BasicResponse) *GzipResponse {
	gr := &GzipResponse{
		BasicResponse: br,
		inner:         br.BodyWriter,
	}
	// Set the Content-Encoding header
	br.Headers.Set("Content-Encoding", "gzip")
	// Assign GzipResponse as the BodyWriter
	br.BodyWriter = gr
	return gr
}
