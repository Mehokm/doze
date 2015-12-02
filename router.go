package rest

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var _ = fmt.Sprint()

var regParam = regexp.MustCompile(`{(\w+)(:\w+)?}`)

const (
	IntParam    = "i"
	StringParam = "a"
)

var regMap = map[string]string{
	IntParam:    `([0-9]+)`,
	StringParam: `([0-9A-Za-z]+)`,
}

type Route struct {
	Path   string
	action map[string]ControllerAction
	params []string
	Params map[string]interface{}
	regex  *regexp.Regexp
}

/*
* with the below comment, maybe remove the abbr for regex and just
* have users supply their own regex.  It would simplify the init()
 */
func (r *Route) init() {
	chunks := strings.Split(r.Path, "/")
	regChunks := make([]string, len(chunks))

	var params []string
	for i, chunk := range chunks {
		if isParam := regParam.MatchString(chunk); isParam {
			trimmed := strings.Trim(chunk, "{}")
			param := trimmed
			paramType := StringParam
			regex := `([^/]+)`

			if cIndex := strings.Index(trimmed, ":"); cIndex != -1 {
				param = trimmed[:cIndex]
				paramType = trimmed[cIndex+1:]

				if reg, valid := regMap[paramType]; valid {
					regex = reg
				}
			}
			params = append(params, param)
			regChunks[i] = regex
		} else {
			regChunks[i] = chunk
		}
	}
	r.regex = regexp.MustCompile(strings.Join(regChunks, "/") + "/?")
	r.params = params
}

/*
* maybe better way to match?  Split req uri and route uri an compare
* if the differences are either something and a param, eg. in { }, then
* run the regex that is in the { }, if any then default regex, against
* the same indexed chunk.  If it passes then move on to the next,
* otherwise its not a match.
 */
func (r *Route) match(test string) bool {
	matches := r.regex.FindStringSubmatch(test)
	if matches != nil && matches[0] == test {
		r.Params = make(map[string]interface{})

		for i, m := range matches[1:] {
			iM, err := strconv.Atoi(m)
			if err == nil {
				r.Params[r.params[i]] = iM
			} else {
				r.Params[r.params[i]] = m
			}
		}
		return true
	}
	return false
}

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
