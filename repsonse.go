package rest

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
)

// ResponseSender is an interface to send response from ControllerActions
type ResponseSender interface {
	Send(w io.Writer)
}

// BasicResponse contains all basic properties for a response
type BasicResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// GzipResponse is a wrapping response to gzip your BasicResponse
type GzipResponse struct {
	br BasicResponse
}

// NewOKJSONResponse returns a BasicResponse tailored for JSON with status code of 200
func NewOKJSONResponse(body interface{}) BasicResponse {
	jr := basicJSONResponse(body)
	jr.StatusCode = http.StatusOK
	return jr
}

// NewCreatedJSONResponse returns a BasicResponse tailored for JSON with status code of 201
func NewCreatedJSONResponse(body interface{}) BasicResponse {
	jr := basicJSONResponse(body)
	jr.StatusCode = http.StatusCreated
	return jr
}

// NewNoContentResponse returns a BasicResponse defaulted for no content
func NewNoContentResponse() BasicResponse {
	return BasicResponse{
		StatusCode: http.StatusNoContent,
		Body:       []byte(http.StatusText(http.StatusNoContent)),
	}
}

// NewNotFoundResponse returns a BasicResponse defaulted for not found
func NewNotFoundResponse() BasicResponse {
	return BasicResponse{
		StatusCode: http.StatusNotFound,
		Body:       []byte(http.StatusText(http.StatusNotFound)),
	}
}

// NewInternalServerErrorResponse returns a BasicResponse defaulted for an internal server error
func NewInternalServerErrorResponse() BasicResponse {
	return BasicResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       []byte(http.StatusText(http.StatusInternalServerError)),
	}
}

func (br BasicResponse) setHeaders(w http.ResponseWriter) {
	for k, v := range br.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(br.StatusCode)
}

// Send writes the BasicResponse body to the http.ResponseWriter
func (br BasicResponse) Send(w io.Writer) {
	if rw, ok := w.(http.ResponseWriter); ok {
		br.setHeaders(rw)
	}

	w.Write(br.Body)
}

func basicJSONResponse(body interface{}) BasicResponse {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	jsonBody, _ := json.Marshal(body)

	br := BasicResponse{}
	br.Headers = headers
	br.Body = jsonBody

	return br
}

// NewGzipResponse creates a new GzipResponse that wraps a BasicResponse that adds Content-Encoding to gzip
func NewGzipResponse(br BasicResponse) GzipResponse {
	br.Headers["Content-Encoding"] = "gzip"

	return GzipResponse{br}
}

// Send creates a gzip writer and writes to the http.ResponseWriter
func (gr GzipResponse) Send(w io.Writer) {
	gr.br.setHeaders(w.(http.ResponseWriter))

	gz := gzip.NewWriter(w)
	defer gz.Close()

	gr.br.Send(gz)
}
