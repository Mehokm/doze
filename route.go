package rest

import (
	"regexp"
	"strconv"
	"strings"
)

var regParam = regexp.MustCompile(`{(\w+)(:\w+)?}`)

const (
	IntParam      = "i"
	AlphaNumParam = "an"
)

var regMap = map[string]string{
	IntParam:      `([0-9]+)`,
	AlphaNumParam: `([0-9A-Za-z]+)`,
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
			regex := `([^/]+)`

			if cIndex := strings.Index(trimmed, ":"); cIndex != -1 {
				param = trimmed[:cIndex]
				regType := trimmed[cIndex+1:]

				if reg, valid := regMap[regType]; valid {
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
