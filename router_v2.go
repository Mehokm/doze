package doze

import (
	"crypto/md5"
	"io"
	"unicode"
)

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
	Children map[rune]*NodeV2
	IsLeaf   bool
	Value    Route
}

func NewNodeV2() *NodeV2 {
	return &NodeV2{
		Children: make(map[rune]*NodeV2),
		IsLeaf:   false,
	}
}

func insert(node *NodeV2, s string, value Route) {
	for _, ch := range s {
		if node.Children[ch] == nil {
			node.Children[ch] = NewNodeV2()
		}
		node = node.Children[ch]
	}

	node.IsLeaf = true
	node.Value = value
}

func search(node *NodeV2, key string) *NodeV2 {
	if key != "" {

		key = key + "$"
	}

	var consuming bool
	var isInt bool = true
	var isAlpha bool = true

	for i, ch := range key {
		if consuming && (ch == '/' || ch == '$') {
			consuming = false

			tmp := node

			skey := key[i:]
			if skey[len(skey)-1] == '$' {
				skey = skey[:len(skey)-1]
			}

			// test for int
			if isInt && node.Children['<'] != nil {
				node = node.Children['<']

				n := search(node, skey)

				if n != nil {
					return n
				}

				node = tmp
			}

			if isAlpha && node.Children['>'] != nil {
				node = node.Children['>']

				n := search(node, skey)

				if n != nil {
					return n
				}

				node = tmp
			}

			// test for wildcard
			if node.Children['*'] != nil {
				node = node.Children['*']

				n := search(node, skey)

				if n != nil {
					return n
				}

				return nil
			}
		} else if node.Children[ch] != nil {
			node = node.Children[ch]
		} else {
			if ch != '$' && len(node.Children) == 0 && node.IsLeaf {
				return nil
			}

			consuming = true

			if !unicode.IsDigit(ch) {
				isInt = false
			}

			if !unicode.IsLetter(ch) {
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
	node := search(r.root, test)

	if node != nil {
		return PatternedRoute{node.Value}, true
	}

	return PatternedRoute{}, false
}
