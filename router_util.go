package rest

import (
	"strconv"
	"strings"
	"unicode"
)

const (
	intParam      = "i"
	alphaParam    = "a"
	alphaNumParam = "an"
)

type pattern struct {
	key     string
	allowed string
}

type param struct {
	name  string
	value interface{}
}

type routeUri struct {
	parts    []string
	patterns map[int]pattern
}

type testUri struct {
	parts  []string
	params map[int]param
}

func newRouteUri(path string) routeUri {
	var parts []string
	var patterns = make(map[int]pattern)

	parts = strings.Split(path, "/")

	routePatternMapperFunc(parts, func(i int, param, pType string) {
		patterns[i] = pattern{param, pType}
	})

	return routeUri{parts, patterns}
}

func newTestUri(path string) testUri {
	var parts []string

	parts = strings.Split(path, "/")

	return testUri{parts, make(map[int]param)}
}

type uriMatcher struct {
	route routeUri
	test  testUri
}

func (um uriMatcher) match() bool {
	if len(um.route.parts) != len(um.test.parts) {
		return false
	}

	for i := 0; i < len(um.test.parts); i++ {
		a := um.route.parts[i]
		b := um.test.parts[i]

		if pattern, ok := um.route.patterns[i]; ok {
			testPart := um.test.parts[i]

			switch pattern.allowed {
			case intParam:
				if _, err := strconv.Atoi(testPart); err != nil {
					return false
				}
			case alphaParam:
				for _, r := range testPart {
					if !unicode.IsLetter(r) {
						return false
					}
				}
			case alphaNumParam:
				for _, r := range testPart {
					if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
						return false
					}
				}
			}

			um.test.params[i] = param{um.route.patterns[i].key, testPart}
		} else if a != b {
			return false
		}
	}

	return true
}

func routePatternMapperFunc(parts []string, fn func(index int, param, pType string)) {
	for i := 0; i < len(parts); i++ {
		part := parts[i]

		if len(part) > 0 && string(part[0]) == "{" && string(part[len(part)-1]) == "}" {
			param := part[1 : len(part)-1]
			var pType string

			if index := strings.Index(part, ":"); index >= 0 {
				param = part[1:index]
				pType = part[index+1 : len(part)-1]
			}

			fn(i, param, pType)
		}
	}
}
