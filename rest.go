package rest

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
)

// ControllerAction is a type for all controller actions
type ControllerAction func(Context) ResponseSender

// Interceptor is a type for adding an intercepting the request before it is processed
type Interceptor func(Context) bool

// Handler implements http.Handler and contains the router and controllers for the REST api
type handler struct {
	router       router
	interceptors []Interceptor
}

// NewHandler returns a new Handler with router initialized
func NewHandler(router router) *handler {
	return &handler{router, make([]Interceptor, 0)}
}

func (h *handler) AddInterceptor(i Interceptor) {
	h.interceptors = append(h.interceptors, i)
}

func (h *handler) invokeInterceptors(c Context) bool {
	result := true
	for i := 0; i < len(h.interceptors) && result; i++ {
		result = h.interceptors[i](c)
	}

	return result
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := h.router.Match(r.URL.Path)
	if route == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	context := Context{
		Request:        r,
		ResponseWriter: w,
		Route:          route,
	}

	if ok := h.invokeInterceptors(context); !ok {
		return
	}

	action, actionExists := route.actions[r.Method]
	if !actionExists {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	result := action(context)

	result.Send(w)
}

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
