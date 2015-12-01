package rest

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var _ = fmt.Sprint()

var regParam = regexp.MustCompile(`{(\w+)(:\w+)?}`)

const (
	PARAM_I = "i"
	PARAM_A = "a"
)

var regMap = map[string]string{
	PARAM_I: `([0-9]+)`,
	PARAM_A: `([0-9A-Za-z]+)`,
}
var typeMap = map[string]reflect.Kind{
	PARAM_I: reflect.Int,
	PARAM_A: reflect.String,
}

type Route struct {
	Path       string
	action     map[string]ControllerAction
	paramTypes map[string]reflect.Kind
	Params     map[string]interface{}
	regex      *regexp.Regexp
}

/*
* with the below comment, maybe remove the abbr for regex and just
* have users supply their own regex.  It would simplify the init()
 */
func (r *Route) init() {
	chunks := strings.Split(r.Path, "/")
	regChunks := make([]string, len(chunks))

	paramTypes := make(map[string]reflect.Kind)
	for i, chunk := range chunks {
		if isParam := regParam.MatchString(chunk); isParam {
			trimmed := strings.Trim(chunk, "{}")
			param := trimmed
			paramType := PARAM_A
			regex := `([^/]+)`

			if cIndex := strings.Index(trimmed, ":"); cIndex != -1 {
				param = trimmed[:cIndex]
				paramType = trimmed[cIndex+1:]
				if reg, valid := regMap[paramType]; valid {
					regex = reg
				}
			}
			paramTypes[param] = typeMap[paramType]
			regChunks[i] = regex
		} else {
			regChunks[i] = chunk
		}
	}
	r.regex = regexp.MustCompile(strings.Join(regChunks, "/") + "/?")
	r.paramTypes = paramTypes
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
		paramKeys := make([]string, len(r.paramTypes))
		index := 0
		for key, _ := range r.paramTypes {
			paramKeys[index] = key
			index++
		}
		for i, m := range matches[1:] {
			switch r.paramTypes[paramKeys[i]] {
			case reflect.Int:
				iM, err := strconv.Atoi(m)
				if err != nil {
					panic(err)
				}
				r.Params[paramKeys[i]] = iM
			default:
				r.Params[paramKeys[i]] = m
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
