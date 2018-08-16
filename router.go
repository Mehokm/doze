package doze

import (
	"regexp"
	"strings"
	"sync"
)

const (
	intParam      = "i"
	alphaParam    = "a"
	alphaNumParam = "an"
)

var regParam = regexp.MustCompile(`{(\w+)(:\w+)?}`)
var regMap = map[string]string{
	intParam:      `([0-9]+)`,
	alphaParam:    `([A-Za-z]+)`,
	alphaNumParam: `([0-9A-Za-z]+)`,
}

type router struct {
	prefix     string
	Routes     map[string]*Route
	routingMap map[*Route]*regexp.Regexp
}

type RouteBuilder struct {
	path      string
	actions   map[string]ActionFunc
	routeName string
}

var (
	routers map[string]router
	lock    sync.RWMutex
)

func init() {
	routers = make(map[string]router)
	routers["default"] = router{"", make(map[string]*Route), make(map[*Route]*regexp.Regexp)}
}

// DefaultRouter returns the router with name "default"
func DefaultRouter() router {
	lock.Lock()
	defer lock.Unlock()

	return routers["default"]
}

// Router returns a router specified by a name
func Router(name string) router {
	lock.Lock()
	defer lock.Unlock()

	// create new router if it doesn't exist
	if _, ok := routers[name]; !ok {
		routers[name] = router{"", make(map[string]*Route), make(map[*Route]*regexp.Regexp)}
	}

	return routers[name]
}

func (ro router) SetPrefix(prefix string) router {
	ro.prefix = prefix

	return ro
}

func (ro router) Prefix() string {
	return ro.prefix
}

// NewRoute returns a wrapper to make a builder for Route
func NewRoute() *RouteBuilder {
	return &RouteBuilder{actions: make(map[string]ActionFunc)}
}

func (rb *RouteBuilder) Name(name string) *RouteBuilder {
	rb.routeName = name

	return rb
}

func (rb *RouteBuilder) For(path string) *RouteBuilder {
	rb.path = path

	return rb
}

func (rb *RouteBuilder) With(method string, action ActionFunc) *RouteBuilder {
	rb.actions[method] = action

	return rb
}

func (rb *RouteBuilder) And(method string, action ActionFunc) *RouteBuilder {
	return rb.With(method, action)
}

func (ro router) RouteMap(rbs ...*RouteBuilder) router {
	for _, RouteBuilder := range rbs {
		route := &Route{
			Path:    ro.prefix + RouteBuilder.path,
			Actions: RouteBuilder.actions,
		}

		ro.initRoute(route)

		if RouteBuilder.routeName == "" {
			ro.Routes[RouteBuilder.path] = route
		}
		ro.Routes[RouteBuilder.routeName] = route
	}

	return ro
}

func (ro router) initRoute(route *Route) {
	toSub := regParam.FindAllStringSubmatch(route.Path, -1)

	regString := route.Path

	if len(toSub) > 0 {
		params := make([]string, len(toSub))

		for i, v := range toSub {
			whole, param, pType, regex := v[0], v[1], v[2], `([^/]+)`

			params[i] = param

			if len(pType) > 1 {
				if r, ok := regMap[pType[1:]]; ok {
					regex = r
				}
			}
			regString = strings.Replace(regString, whole, regex, -1)
		}
		ro.SetParamNames(route, params)
	}

	ro.routingMap[route] = regexp.MustCompile(regString + "/?")
}

func (ro router) Get(name string) *Route {
	return ro.Routes[name]
}

func (ro router) Match(test string) *Route {
	for route, regex := range ro.routingMap {
		matches := regex.FindStringSubmatch(test)
		if matches != nil && matches[0] == test {
			values := make([]interface{}, len(matches[1:]))

			for i, m := range matches[1:] {
				values[i] = m
			}
			ro.SetParamValues(route, values)

			return route
		}
	}

	return nil
}

func (ro router) SetParamNames(r *Route, pn []string) {
	r.ParamNames = pn
}

func (ro router) SetParamValues(r *Route, pv []interface{}) {
	r.ParamValues = pv
}

func (ro router) SetActions(r *Route, a map[string]ActionFunc) {
	r.Actions = a
}
