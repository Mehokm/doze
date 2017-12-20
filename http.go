package doze

import "net/http"

// ResponseWriter is a wrapper for http.ResponseWriter which includes extra properties
// to keep track of current response
type ResponseWriter struct {
	http.ResponseWriter
	Size       int
	StatusCode int
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)

	if err == nil {
		rw.Size += size
	}

	return size, err
}

// WriteHeader is overidding http.ResponseWriter in order to capture the
// StatusCode of the request
func (rw *ResponseWriter) WriteHeader(i int) {
	rw.ResponseWriter.WriteHeader(i)

	rw.StatusCode = i
}

// Written returns whether or not the current response has been written to
func (rw *ResponseWriter) Written() bool {
	return rw.Size > 0
}

// Request is a wrapper to *http.Request which has a Data property to allow
// key-value pair store to share data along the request
type Request struct {
	*http.Request
	Data requestData
}

type requestData struct {
	data map[interface{}]interface{}
}

func (rd requestData) Set(key, value interface{}) {
	rd.data[key] = value
}

func (rd requestData) Get(key interface{}) interface{} {
	if value, ok := rd.data[key]; ok {
		return value
	}

	return nil
}
