package rest

import "net/http"

const (
	MethodGET    = "GET"
	MethodPOST   = "POST"
	MethodPUT    = "PUT"
	MethodDELETE = "DELETE"
)

type ResponseWriter struct {
	http.ResponseWriter
	Size int
}

func (rw *ResponseWriter) Written() bool {
	return rw.Size > 0
}

// ControllerAction is a type for all controller actions
type ControllerAction func(Context) ResponseSender

// Interceptor is a type for adding an intercepting the request before it is processed
type Interceptor func(Context) bool

type Middleware func(Context)

// Handler implements http.Handler and contains the router and controllers for the REST api
type handler struct {
	router       Routable
	interceptors []Interceptor
	middlewares  []Middleware
}

// NewHandler returns a new Handler with router initialized
func NewHandler(r Routable) *handler {
	return &handler{r, make([]Interceptor, 0), make([]Middleware, 0)}
}

func (h *handler) Intercept(i Interceptor) {
	h.interceptors = append(h.interceptors, i)
}

func (h *handler) Use(m Middleware) {
	h.middlewares = append(h.middlewares, m)
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

	action, actionExists := route.actions[r.Method]
	if !actionExists {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	context := Context{
		Request:        r,
		ResponseWriter: &ResponseWriter{w, 0},
		Route:          route,
		middlewares:    h.middlewares,
		action:         action,
	}

	if ok := h.invokeInterceptors(context); !ok {
		return
	}

	context.run()

	// result := action(context)
	//
	// result.Send(w)
}
