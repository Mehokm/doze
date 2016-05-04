package rest

type router struct {
	Routes map[string]*Route
}

type routeBuilder struct {
	path       string
	controller interface{}
	action     map[string]ControllerAction
	routeName  string
}

var defaultRouter router

func init() {
	defaultRouter = router{make(map[string]*Route)}
}

func DefaultRouter() router {
	return defaultRouter
}

func NewRoute() *routeBuilder {
	return &routeBuilder{action: make(map[string]ControllerAction)}
}

func (rb *routeBuilder) Named(name string) *routeBuilder {
	rb.routeName = name
	return rb
}

func (rb *routeBuilder) For(path string) *routeBuilder {
	rb.path = path
	return rb
}

func (rb *routeBuilder) With(method string, action ControllerAction) *routeBuilder {
	rb.action[method] = action
	return rb
}

func (rb *routeBuilder) And(method string, action ControllerAction) *routeBuilder {
	return rb.With(method, action)
}

func (ro router) RouteMap(rbs ...*routeBuilder) router {
	for _, routeBuilder := range rbs {
		route := &Route{
			Path:   routeBuilder.path,
			action: routeBuilder.action,
		}
		route.init()

		if routeBuilder.routeName == "" {
			ro.Routes[routeBuilder.path] = route
		}
		ro.Routes[routeBuilder.routeName] = route
	}
	return ro
}

func (ro router) GetRoute(name string) *Route {
	return ro.Routes[name]
}

func (ro router) Match(test string) *Route {
	for _, route := range ro.Routes {
		if route.match(test) {
			return route
		}
	}
	return nil
}
