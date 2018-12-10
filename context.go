package doze

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Context will contain all information about the current request scoped context
type Context struct {
	Request        *http.Request
	ResponseWriter *ResponseWriter
	Route          *Route
}

// Set puts a value on the current context.Context by key
func (c *Context) Set(key, value interface{}) {
	ctx := context.WithValue(c.Request.Context(), key, value)

	c.Request = c.Request.WithContext(ctx)
}

// Value returns the value by key on the current context.Context
func (c *Context) Value(key interface{}) interface{} {
	return c.Request.Context().Value(key)
}

// FormData returns data related to the request from GET, POST, or PUT
func (c *Context) FormData() url.Values {
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
func (c *Context) BindJSONEntity(i interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(&i)
}
