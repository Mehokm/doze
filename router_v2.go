package doze

import (
	"crypto/md5"
	"io"
	"unicode"
)

const MaxSize = 87
const END = '$'
const INT = '<' - END
const ALPHA = '>' - END
const WILDCARD = '*' - END
const SLASH = '/' - END

type routerV2 struct {
	routes map[string]Route
	root   *NodeV2
}

func NewRouterV2() routerV2 {
	return routerV2{
		routes: make(map[string]Route),
		root:   NewNodeV2(),
	}
}

func (r routerV2) GET(path string, fn ActionFunc) {
	h := md5.New()
	io.WriteString(h, path)

	name := h.Sum(nil)

	route := NewRoute().For(path).With("GET", fn)

	r.routes[string(name)] = route

	insert(r.root, path, route)
}

type NodeV2 struct {
	Children    [MaxSize]*NodeV2
	HasChildren bool
	IsLeaf      bool
	Value       Route
}

func NewNodeV2() *NodeV2 {
	return &NodeV2{
		IsLeaf: false,
	}
}

func isEmpty(arr [MaxSize]*NodeV2) bool {
	for _, n := range arr {
		if n != nil {
			return false
		}
	}

	return true
}

func insert(node *NodeV2, s string, value Route) {
	for _, ch := range s {
		ch = ch - END

		if node.Children[ch] == nil {
			node.Children[ch] = NewNodeV2()
		}
		node = node.Children[ch]
	}

	node.HasChildren = !isEmpty(node.Children)
	node.IsLeaf = true
	node.Value = value
}

func search(node *NodeV2, key []byte) *NodeV2 {
	if key != nil {
		key = append([]byte(key), '$')
	}

	var consuming bool
	var isInt bool = true
	var isAlpha bool = true

	for i, ch := range key {
		c := ch - END

		if consuming && (c == SLASH || ch == END) {
			var valid bool = true

			consuming = false

			tmp := node

			skey := key[i:]
			if skey[len(skey)-1] == END {
				skey = skey[:len(skey)-1]
			}

			// test for int
			if isInt && node.Children[INT] != nil {
				node = node.Children[INT]

				found := search(node, skey)

				if found != nil {
					return found
				}

				valid = false

				node = tmp
			}

			if isAlpha && node.Children[ALPHA] != nil {
				node = node.Children[ALPHA]

				found := search(node, skey)

				if found != nil {
					return found
				}

				valid = false

				node = tmp
			}

			// test for wildcard
			if node.Children[WILDCARD] != nil {
				node = node.Children[WILDCARD]

				found := search(node, skey)

				if found != nil {
					return found
				}

				valid = false
			}

			if !valid {
				return nil
			}
		} else if node.Children[c] != nil {
			node = node.Children[c]
		} else {
			if ch != END && !node.HasChildren && node.IsLeaf {
				return nil
			}

			consuming = true

			if !unicode.IsDigit(rune(ch)) {
				isInt = false
			}

			if !unicode.IsLetter(rune(ch)) {
				isAlpha = false
			}
		}
	}

	if node != nil && node.IsLeaf {
		return node
	}

	return nil
}

func (r routerV2) Get(name string) PatternedRoute {
	return PatternedRoute{}
}

func (r routerV2) Match(test string) (PatternedRoute, bool) {
	node := search(r.root, []byte(test))

	if node != nil {
		return PatternedRoute{node.Value}, true
	}

	return PatternedRoute{}, false
}
