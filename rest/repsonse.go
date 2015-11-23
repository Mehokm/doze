package rest

import (
	"encoding/json"
	"net/http"
)

type Response interface {
	Send(w http.ResponseWriter)
}

type JsonResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       interface{}
}

func NewOKJsonResponse(body interface{}) JsonResponse {
	jr := basicJsonResponse()
	jr.StatusCode = http.StatusOK
	jr.Body = body
	return jr
}

func NewCreatedJsonResponse(body interface{}) JsonResponse {
	jr := basicJsonResponse()
	jr.StatusCode = http.StatusCreated
	jr.Body = body
	return jr
}

func NewNoContentJsonResponse() JsonResponse {
	jr := basicJsonResponse()
	jr.StatusCode = http.StatusNoContent
	jr.Body = http.StatusText(http.StatusNoContent)
	return jr
}

func NewNotFoundJsonResponse() JsonResponse {
	jr := basicJsonResponse()
	jr.StatusCode = http.StatusNotFound
	jr.Body = http.StatusText(http.StatusNotFound)
	return jr
}

func (jr JsonResponse) Send(w http.ResponseWriter) {
	for k, v := range jr.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(jr.StatusCode)

	jsonBody, _ := json.Marshal(jr.Body)
	w.Write(jsonBody)
}

func basicJsonResponse() JsonResponse {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	jr := JsonResponse{}
	jr.Headers = headers
	return jr
}
