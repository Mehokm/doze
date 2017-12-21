package doze

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type Context struct {
	Request        *http.Request
	ResponseWriter *ResponseWriter
	Route          *Route
	middlewares    []Middleware
	mIndex         int
	action         Action
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

// Next calls the next middleware in the chain
func (c *Context) Next() {
	c.mIndex++

	c.run()
}

func (c *Context) run() {
	if c.ResponseWriter.Written() {
		return
	}

	if c.mIndex < len(c.middlewares) {
		c.middlewares[c.mIndex](c)

		c.mIndex++
	} else if c.mIndex == len(c.middlewares) {
		result := c.action(c)

		if result != nil {
			_, err := result.Send(c.ResponseWriter)

			if err != nil {
				panic(err)
			}
		}
	}
}
