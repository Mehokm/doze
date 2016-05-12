package rest

type Routable interface {
	Get(string) *Route
	Match(string) *Route
}

type router struct {
	prefix string
	Routes map[string]*Route
}

type routeBuilder struct {
	path      string
	actions   map[string]ControllerAction
	routeName string
}

var routers map[string]router

func init() {
	routers = make(map[string]router)
	routers["default"] = router{"", make(map[string]*Route)}
}

func NewRouter(name string) router {
	routers[name] = router{"", make(map[string]*Route)}
	return routers[name]
}

func DefaultRouter() router {
	return routers["default"]
}

func Router(name string) router {
	return routers[name]
}

func (r router) Prefix(prefix string) router {
	r.prefix = prefix
	return r
}

func NewRoute() *routeBuilder {
	return &routeBuilder{actions: make(map[string]ControllerAction)}
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
	rb.actions[method] = action
	return rb
}

func (rb *routeBuilder) And(method string, action ControllerAction) *routeBuilder {
	return rb.With(method, action)
}

func (ro router) RouteMap(rbs ...*routeBuilder) router {
	for _, routeBuilder := range rbs {
		route := &Route{
			path:    ro.prefix + routeBuilder.path,
			actions: routeBuilder.actions,
		}
		route.init()

		if routeBuilder.routeName == "" {
			ro.Routes[routeBuilder.path] = route
		}
		ro.Routes[routeBuilder.routeName] = route
	}

	return ro
}

func (ro router) Get(name string) *Route {
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
