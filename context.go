package rest

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Route          *Route
}

// FormData returns data related to the request from GET, POST, or PUT
func (c Context) FormData() url.Values {
	c.Request.ParseForm()
	switch c.Request.Method {
	case "POST":
		fallthrough
	case "PUT":
		return c.Request.PostForm
	default:
		return c.Request.Form
	}
}

// BindJSONEntity binds the JSON body from the request to an interface{}
func (c Context) BindJSONEntity(i interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(&i)
}
