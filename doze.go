package doze

import (
	"net/http"
)

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
	SetActions(*Route, map[string]ActionFunc)
}

// Action is a type for all controller actions
type ActionFunc func(*Context) ResponseSender

// Handler implements http.Handler and contains the router and controllers for the REST api
type Handler struct {
	Router          Routeable
	middlewareChain *middlewareChain
}

// NewHandler returns a new Handler with router initialized
func NewHandler(r Routeable) *Handler {
	return &Handler{Router: r, middlewareChain: new(middlewareChain)}
}

func (h *Handler) Pattern() string {
	return h.Router.(router).prefix + "/"
}

func (h *Handler) Use(mf MiddlewareFunc) {
	h.middlewareChain.add(mf)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := h.Router.Match(r.URL.Path)
	if route == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	action, actionExists := route.Actions[r.Method]
	if !actionExists {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	context := &Context{
		Request:        r,
		ResponseWriter: &ResponseWriter{w, 0, 0},
		Route:          route,
	}

	h.middlewareChain.action = action

	h.middlewareChain.run(context)
	return
}
