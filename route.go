package rest

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	intParam      = "i"
	alphaNumParam = "an"
)

var regParam = regexp.MustCompile(`{(\w+)(:\w+)?}`)
var regMap = map[string]string{
	intParam:      `([0-9]+)`,
	alphaNumParam: `([0-9A-Za-z]+)`,
}

type Route struct {
	Path        string
	Actions     map[string]Action
	ParamNames  []string
	ParamValues []interface{}
}

// Params returns a key-value pair containing the route parameters defined in the
// route path.  ParamNames should alwyas go 1-1 to the ParamValues, otherwise you
// will have a bad time
func (r *Route) Params() map[string]interface{} {
	pv := make(map[string]interface{})

	for i, v := range r.ParamValues {
		pv[r.ParamNames[i]] = v
		if n, err := strconv.Atoi(v.(string)); err == nil {
			pv[r.ParamNames[i]] = n
		}
	}
	return pv
}

// Build returns the route path with route parameters replaced with values from the
// passed in map
func (r *Route) Build(m map[string]interface{}) (string, error) {
	if len(r.ParamNames) != len(m) {
		return "", fmt.Errorf("wrong number of parameters: %v given, %v required", len(m), len(r.ParamNames))
	}

	s := r.Path
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
