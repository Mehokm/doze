package doze

import (
	"net/http"
)

// Routeable is an interface which allows you to create your own router
// * Get(string) *Route returns the route by route name
// * Match(string) *Route takes a URI and returns a *Route it matches.  If it does not
//   it returns nil
type Routeable interface {
	Get(string) PatternedRoute
	Match(string) (PatternedRoute, bool)
}

// ActionFunc is a type that is a function to be used as a controller action
type ActionFunc func(*Context) ResponseSender

// Handler implements http.Handler and contains the router and controllers for the REST api
type Handler struct {
	router     Routeable
	middleware []MiddlewareFunc
}

// NewHandler returns a new Handler with router initialized
func NewHandler(r Routeable) *Handler {
	return &Handler{router: r}
}

// Use applies a MiddlewareFunc to be executed in the request chain
func (h *Handler) Use(mf MiddlewareFunc) {
	h.middleware = append(h.middleware, mf)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, matched := h.router.Match(r.URL.Path)
	if !matched {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	action, actionExists := route.Actions()[r.Method]
	if !actionExists {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	context := &Context{
		Request:        r,
		ResponseWriter: &ResponseWriter{w, 0, 0},
		Route:          route,
	}

	mwc := &middlewareChain{action: action}
	for _, mw := range h.middleware {
		mwc.add(mw)
	}

	mwc.run(context)
	return
}
