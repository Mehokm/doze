package rest

import (
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
	action         ControllerAction
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

// EXPERIMENT

// Next calls the next middleware in the chain
func (c Context) Next() {
	c.mIndex++

	c.run()
}

func (c Context) run() {
	for {
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
}
