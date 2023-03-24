package doze

import (
	"crypto/md5"
	"fmt"
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
	routes    map[string]Route
	root      *NodeV2
	wildcards map[int]bool
}

func NewRouterV2() routerV2 {
	return routerV2{
		routes:    make(map[string]Route),
		root:      NewNodeV2(),
		wildcards: make(map[int]bool),	// THIS NEEDS TO BE A MAP OF WILDCARD MAPS INDEXED BY THE PATH SINCE THE ROUTER HAS MULTIPLE ROUTES WHICH COULD MESS UP THE LOGIC
	}
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

type ParamType int

type Param struct {
	Pos   int
	Value interface{}
	Type  ParamType
}

func (r *routerV2) GET(path string, fn ActionFunc) {
	h := md5.New()
	io.WriteString(h, path)

	name := h.Sum(nil)

	route := NewRoute().For(path).With("GET", fn)

	r.routes[string(name)] = route

	insert(r.root, r.wildcards, path, route)
}

type Key struct {
	KeyPos int
	Val    string
}

func (k Key) Consume() Key {
	return Key{
		KeyPos: k.KeyPos,
		Val:    k.Val[:k.KeyPos] + k.Val[k.KeyPos+1:],
	}
}

func (k Key) Append(ch rune) Key {
	return Key{
		KeyPos: k.KeyPos + 1,
		Val:    k.Val + string(ch),
	}
}

func (k Key) InsertAt(ch rune, index int) Key {
	return Key{
		KeyPos: k.KeyPos,
		Val:    k.Val[:index] + string(ch) + k.Val[index+1:],
	}
}

func (k Key) Next() Key {
	return Key{
		KeyPos: k.KeyPos + 1,
		Val:    k.Val,
	}
}

func (k Key) Current() rune {
	return rune(k.Val[k.KeyPos])
}

func (k Key) HasNext() bool {
	return k.KeyPos < len(k.Val)
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

func insert(node *NodeV2, wildcards map[int]bool, s string, value Route) {
	for i, ch := range s {
		ch = ch - END

		if node.Children[ch] == nil {
			node.Children[ch] = NewNodeV2()

			if ch == WILDCARD {
				wildcards[i] = true	//MAYBE WE SHOULD STORE THE PARAMS AS THE VALUE OF THE WILDCARD SO WE HAVE IT
			}
		}
		node = node.Children[ch]
	}

	node.HasChildren = !isEmpty(node.Children)
	node.IsLeaf = true
	node.Value = value
}

func search2(node *NodeV2, wildcards map[int]bool, key Key) *NodeV2 {
start:
	for key.HasNext() {
		ch := key.Current()
		ch = ch - END

		if node.Children[ch] == nil {
			if wildcards[key.KeyPos] {
				node = node.Children[WILDCARD]

				// consume key until we find a slash or end
				for {
					if key.KeyPos+1 >= len(key.Val) {
						if node.IsLeaf {
							return node
						}
					} else if key.Val[key.KeyPos+1]-END == SLASH {
						key = key.InsertAt('*', key.KeyPos)
						key = key.Next()
						goto start
					}
					key = key.Consume()
				}
			}

			return nil
		}

		key = key.Next()
		node = node.Children[ch]
	}

	if node != nil && node.IsLeaf {
		return node
	}

	return nil
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

func parseRouteParams(route string) map[int]string {
	// need to make param type with pos, value, and type

	params := make(map[int]string)
	bstack := make([]rune, 0)
	param := make([]rune, 0)

	var chunk int
	for _, ch := range route {
		if ch == '/' {
			chunk++
		}

		if ch == '}' && peek(bstack) == '{' {
			params[chunk] = string(param)

			param = make([]rune, 0)
			pop(&bstack)
		}

		if peek(bstack) == '{' {
			param = append(param, ch)
		}

		if ch == '{' {
			if len(bstack) > 0 {
				param = make([]rune, 0)
			}
			push(&bstack, ch)
		}
	}

	for _, p := range params {
		n := len(p)

		if n > 2 && p[n-2] == ':' {
			name := p[:n-2]
			t := p[n-1:]

			fmt.Println(name, t)
		}
	}

	return params
}

func push(stack *[]rune, item rune) {
	*stack = append(*stack, item)
}

func pop(stack *[]rune) rune {
	if len(*stack) > 0 {
		item := (*stack)[len(*stack)-1]
		*stack = (*stack)[:len(*stack)-1]

		return item
	}

	return -1
}

func peek(stack []rune) rune {
	if len(stack) > 0 {
		return stack[len(stack)-1]
	}
	return -1
}
