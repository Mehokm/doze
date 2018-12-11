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

type RestRouter struct {
	prefix     string
	routes     map[string]Route
	routingMap map[Route]*regexp.Regexp
}

var (
	routers map[string]RestRouter
	lock    sync.RWMutex
)

func init() {
	routers = make(map[string]RestRouter)
}

// NewRouter returns a router specified by a name
func Router(name string) RestRouter {
	lock.Lock()
	defer lock.Unlock()

	// create new router if it doesn't exist
	if _, ok := routers[name]; !ok {
		routers[name] = RestRouter{"", make(map[string]Route), make(map[Route]*regexp.Regexp)}
	}

	return routers[name]
}

func (ro RestRouter) SetPrefix(prefix string) RestRouter {
	ro.prefix = prefix

	return ro
}

func (ro RestRouter) Prefix() string {
	return ro.prefix
}

// NewRoute returns a new *DozeRoute
func NewRoute() *DozeRoute {
	return &DozeRoute{actions: make(map[string]ActionFunc)}
}

func (r *DozeRoute) Named(name string) *DozeRoute {
	r.SetName(name)

	return r
}

func (r *DozeRoute) For(path string) *DozeRoute {
	r.SetPath(path)

	return r
}

func (r *DozeRoute) With(method string, action ActionFunc) *DozeRoute {
	actions := r.Actions()
	actions[method] = action
	r.SetActions(actions)

	return r
}

func (r *DozeRoute) And(method string, action ActionFunc) *DozeRoute {
	return r.With(method, action)
}

func (ro RestRouter) Add(route Route) {
	route.SetPath(ro.Prefix() + route.Path())

	initRoute(ro, route)

	if route.Name() != "" {
		ro.routes[route.Name()] = route
	} else {
		ro.routes[route.Path()] = route
	}
}

func initRoute(router RestRouter, route Route) {
	toSub := regParam.FindAllStringSubmatch(route.Path(), -1)

	regString := route.Path()

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

		route.SetParamNames(params)
	}

	router.routingMap[route] = regexp.MustCompile(regString + "/?")
}

func (ro RestRouter) Get(name string) PatternedRoute {
	return PatternedRoute{ro.routes[name]}
}

func (ro RestRouter) Match(test string) (PatternedRoute, bool) {
	for route, regex := range ro.routingMap {
		matches := regex.FindStringSubmatch(test)
		if matches != nil && matches[0] == test {
			values := make([]interface{}, len(matches[1:]))

			for i, m := range matches[1:] {
				values[i] = m
			}

			route.SetParamValues(values)

			return PatternedRoute{route}, true
		}
	}

	return PatternedRoute{}, false
}
