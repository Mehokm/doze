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
	anythingParam = ""
)

type Pattern struct {
	key     string
	allowed string
}

type Param struct {
	name  string
	value interface{}
}

type RouteUri struct {
	parts    []string
	patterns map[int]Pattern
}

type TestUri struct {
	parts  []string
	params map[int]Param
}

func NewRouteUri(path string) RouteUri {
	var parts []string
	var patterns = make(map[int]Pattern)

	parts = strings.Split(path, "/")

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		if len(part) > 0 && string(part[0]) == "{" && string(part[len(part)-1]) == "}" {
			key := part[1 : len(part)-1]
			var allowed string

			if index := strings.Index(part, ":"); index >= 0 {
				key = part[1:index]
				allowed = part[index+1 : len(part)-1]
			}

			patterns[i] = Pattern{key, allowed}
		}
	}

	return RouteUri{parts, patterns}
}

func NewTestUri(path string) TestUri {
	var parts []string

	parts = strings.Split(path, "/")

	return TestUri{parts, make(map[int]Param)}
}

type UriMatcher struct {
	route RouteUri
	test  TestUri
}

func (um UriMatcher) match() bool {
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

			um.test.params[i] = Param{um.route.patterns[i].key, testPart}
		} else if a != b {
			return false
		}
	}

	return true
}
