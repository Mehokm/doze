package doze

import (
	"fmt"
	"regexp"
	"strconv"
)

type DozeRoute struct {
	name        string
	path        string
	actions     map[string]ActionFunc
	paramNames  []string
	paramValues []interface{}
}

type Route interface {
	Name() string
	Path() string
	Actions() map[string]ActionFunc
	ParamNames() []string
	ParamValues() []interface{}

	SetName(string)
	SetPath(string)
	SetActions(map[string]ActionFunc)
	SetParamNames([]string)
	SetParamValues([]interface{})
}

func (r *DozeRoute) Name() string {
	return r.name
}

func (r *DozeRoute) Path() string {
	return r.path
}

func (r *DozeRoute) Actions() map[string]ActionFunc {
	return r.actions
}

func (r *DozeRoute) ParamNames() []string {
	return r.paramNames
}

func (r *DozeRoute) ParamValues() []interface{} {
	return r.paramValues
}

func (r *DozeRoute) SetName(name string) {
	r.name = name
}

func (r *DozeRoute) SetPath(path string) {
	r.path = path
}

func (r *DozeRoute) SetActions(actions map[string]ActionFunc) {
	r.actions = actions
}

func (r *DozeRoute) SetParamNames(paramNames []string) {
	r.paramNames = paramNames
}

func (r *DozeRoute) SetParamValues(paramValues []interface{}) {
	r.paramValues = paramValues
}

type PatternedRoute struct {
	Route
}

// Params returns a key-value pair containing the route parameters defined in the
// route path.  ParamNames should alwyas go 1-1 to the ParamValues, otherwise you
// will have a bad time
func (r PatternedRoute) Params() map[string]interface{} {
	pv := make(map[string]interface{})

	paramNames := r.ParamNames()
	for i, v := range r.ParamValues() {
		pv[paramNames[i]] = v
		if n, err := strconv.Atoi(v.(string)); err == nil {
			pv[paramNames[i]] = n
		}
	}
	return pv
}

// Build returns the route path with route parameters replaced with values from the
// passed in map
func (r PatternedRoute) Build(m map[string]interface{}) (string, error) {
	if len(r.ParamNames()) != len(m) {
		return "", fmt.Errorf("wrong number of parameters: %v given, %v required", len(m), len(r.ParamNames()))
	}

	s := r.Path()
	for p, v := range m {
		reg := regexp.MustCompile(fmt.Sprintf(`{%v(:\w+)?}`, p))

		if !reg.MatchString(s) {
			return "", fmt.Errorf("parameter not valid: %v", p)
		}

		switch v.(type) {
		case int:
			i := strconv.Itoa(v.(int))
			s = reg.ReplaceAllString(s, i)
		default:
			s = reg.ReplaceAllString(s, v.(string))
		}

	}
	return s, nil
}
