package rest

import "net/http"

// Routeable is an interface which allows you to create your own router
// * Get(string) *Route returns the route by route name
// * Match(string) *Route takes a URI and returns a *Route it matches.  If it does not
//   it returns nil
// * SetParamNames sets the route parameter names to a route.  Ex. /api/foo/{id}, id is a param
// * SetParamValues sets the value to match the route parameter.  Ex. /api/foo/i, id is the param, and 1 is the value
// * SetActions set the actions for that given route.  The map key is the method.  Ex. map["GET"]Action
type Routeable interface {
	Get(string) *Route
	Match(string) *Route
	SetParamNames(*Route, []string)
	SetParamValues(*Route, []interface{})
	SetActions(*Route, map[string]Action)
}

// Action is a type for all controller actions
type Action func(Context) ResponseSender

// Interceptor is a type for adding an intercepting the request before it is processed
type Interceptor func(Context) bool

// Middleware is a type for adding middleware for the request
type Middleware func(Context)

// Handler implements http.Handler and contains the router and controllers for the REST api
type handler struct {
	router       Routeable
	interceptors []Interceptor
	middlewares  []Middleware
}

// NewHandler returns a new Handler with router initialized
func NewHandler(r Routeable) *handler {
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

	action, actionExists := route.Actions[r.Method]
	if !actionExists {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	context := Context{
		Request:        Request{r, requestData{make(map[interface{}]interface{})}},
		ResponseWriter: &ResponseWriter{w, 0, 0},
		Route:          route,
		middlewares:    h.middlewares,
		action:         action,
	}

	if ok := h.invokeInterceptors(context); !ok {
		// maybe check to see if response and header/status has been written
		// if not, then probably should do something
		return
	}

	context.run()

	return
}
