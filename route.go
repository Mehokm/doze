package rest

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	IntParam      = "i"
	AlphaNumParam = "an"
)

var regParam = regexp.MustCompile(`{(\w+)(:\w+)?}`)
var regMap = map[string]string{
	IntParam:      `([0-9]+)`,
	AlphaNumParam: `([0-9A-Za-z]+)`,
}

type Route struct {
	Path        string
	actions     map[string]Action
	ParamNames  []string
	ParamValues []interface{}
}

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
