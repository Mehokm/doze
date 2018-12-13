package doze

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

	var i int
	regstrs := make([]string, len(m))

	for p := range m {
		regstrs[i] = fmt.Sprintf(`{(%v)(?::\w+)?}`, p)

		i++
	}

	reg := regexp.MustCompile(strings.Join(regstrs, "|"))

	result := reg.ReplaceAllStringFunc(s, func(str string) string {
		i := strings.Index(str, ":")

		var param string
		if i >= 0 { // will cause issue if param is formatted correctly, but that shouldn't happen
			param = str[1:i]
		} else {
			param = str[1 : len(str)-1]
		}

		value, ok := m[param]

		if !ok {
			return ""
		}

		switch value.(type) {
		case int:
			return strconv.Itoa(value.(int))
		case float32:
		case float64:
			return strconv.FormatFloat(value.(float64), 'f', -1, 64)
		case string:
			return value.(string)
		}

		return ""
	})

	return result, nil
}
