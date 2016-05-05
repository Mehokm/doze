package rest

import (
	"regexp"
	"strconv"
	"strings"
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
	path        string
	actions     map[string]ControllerAction
	params      []string
	paramValues map[string]interface{}
	regex       *regexp.Regexp
}

func (r *Route) init() {
	toSub := regParam.FindAllStringSubmatch(r.path, -1)

	regString := r.path

	if len(toSub) > 0 {
		r.params = make([]string, len(toSub))

		for i, v := range toSub {
			whole, param, pType, regex := v[0], v[1], v[2], `([^/]+)`

			r.params[i] = param

			if r, ok := regMap[pType[1:]]; ok {
				regex = r
			}
			regString = strings.Replace(regString, whole, regex, -1)
		}
	}
	r.regex = regexp.MustCompile(regString + "/?")
}

func (r *Route) match(test string) bool {
	matches := r.regex.FindStringSubmatch(test)
	if matches != nil && matches[0] == test {
		r.paramValues = make(map[string]interface{})

		for i, m := range matches[1:] {
			r.paramValues[r.params[i]] = m
		}

		return true
	}

	return false
}

func (r *Route) Params() map[string]interface{} {
	pv := make(map[string]interface{})

	for i, v := range r.paramValues {
		pv[i] = v
		if n, err := strconv.Atoi(v.(string)); err == nil {
			pv[i] = n
		}
	}

	return pv
}
