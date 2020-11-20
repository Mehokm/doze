package doze

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterMatchV2(t *testing.T) {
	r2 := NewRouterV2()

	routes := []string{
		"/foo",
		"/foo/bar",
		"/foo/bar/*",
		"/foo/bar/*/baz",
		"/oof",
		"/oof/rab",
		"/oof/rab/*",
		"/oof/rab/*/baz",
		"/f/*",
		"/f/*/b/c",
		"/f/</b/c/<",
		"/f/</b/c/*/d/>",
		"/foo/bar/*/baz/*/oof/*/rab/*",
		"/foo/bar/*/baz/*/oof/*/rab/*/zab",
		"/foo/bar/</baz/</oof/</rab/</zab/<",
		"/foo/bar/*/baz/*/oof/*/rab/*/zab/*",
	}

	for _, route := range routes {
		r2.GET(route, func(c *Context) ResponseSender {
			return nil
		})
	}

	testcases := []struct {
		test        string
		shouldMatch bool
	}{
		{
			test:        "/foo",
			shouldMatch: true,
		},
		{
			test:        "/fo",
			shouldMatch: false,
		},
		{
			test:        "/foo/bar",
			shouldMatch: true,
		},
		{
			test:        "/foo/ba",
			shouldMatch: false,
		},
		{
			test:        "/foo/bar/asdf",
			shouldMatch: true,
		},
		{
			test:        "/foo/bar/1234",
			shouldMatch: true,
		},
		{
			test:        "/foo/bar/asdf/baz",
			shouldMatch: true,
		},
		{
			test:        "/foo/bar/1234/baz",
			shouldMatch: true,
		},
		{
			test:        "/foo/bar/asdf/ba",
			shouldMatch: false,
		},
		{
			test:        "/f/asdf",
			shouldMatch: true,
		},
		{
			test:        "/f/1234",
			shouldMatch: true,
		},
		{
			test:        "/f/1234/b",
			shouldMatch: false,
		},
		{
			test:        "/f/1234/b/c",
			shouldMatch: true,
		},
		{
			test:        "/f/asdf/b/c",
			shouldMatch: true,
		},
		{
			test:        "/f/1234/b/c/4321",
			shouldMatch: true,
		},
		{
			test:        "/f/asdf/b/c/4321",
			shouldMatch: false,
		},
		{
			test:        "/f/asdf/b/c/4321/d/1234",
			shouldMatch: false,
		},
		{
			test:        "/f/1234/b/c/4321/d/ghjk",
			shouldMatch: true,
		},
		{
			test:        "/foo/bar/A/baz/B/oof/C/rab/D",
			shouldMatch: true,
		},
		{
			test:        "/foo/bar/A/baz/B/oof/C/rab",
			shouldMatch: false,
		},
		{
			test:        "/foo/bar/A/baz/B/oof/C/rab/D/zab",
			shouldMatch: true,
		},
		{
			test:        "/foo/bar/A/baz/B/oof/C/rab/D/zab/E",
			shouldMatch: true,
		},
		{
			test:        "/foo/bar/A/baz/B/oof/C/rab/D/zab/E/F",
			shouldMatch: false,
		},
		{
			test:        "/foo/bar/1/baz/2/oof/3/rab/4/zab/5",
			shouldMatch: true,
		},
		{
			test:        "/foo/bar/1/baz/2/oof/3/rab/4/zab/E",
			shouldMatch: true,
		},
		{
			test:        "/foo/bar/1/baz/2/oof/A/rab/B/zab/C/d",
			shouldMatch: false,
		},
	}

	for _, tt := range testcases {
		t.Run("_"+tt.test, func(t *testing.T) {
			_, matched := r2.Match(tt.test)

			if matched != tt.shouldMatch {
				t.Errorf("got %v, want %v", matched, tt.shouldMatch)
			}
		})
	}
}

func BenchmarkRouterMatch3(b *testing.B) {
	b.StopTimer()

	// wildcard '*'
	// int '<'
	// alpha '>'

	routes := []string{
		"/*/*/*/*/*/*/*/*/*/*/*/*/*/*/*/*/*/*/*/*",
		"/foo",
		"/foo/bar",
		"/foo/bar/*",
		"/foo/bar/*/baz",
		"/oof",
		"/oof/rab",
		"/oof/rab/*",
		"/oof/rab/*/baz",
		"/f/*",
		"/f/*/b/c",
		"/f/</b/c/<",
		"/f/</b/c/*/d/>",
		"/foo/bar/*/baz/*/oof/*/rab/*",
		"/foo/bar/*/baz/*/oof/*/rab/*/zab",
		"/foo/bar/</baz/</oof/</rab/</zab/<",
		"/foo/bar/*/baz/*/oof/*/rab/*/zab/>",
	}

	test := "/a/b/c/d/e/a/b/c/d/e/a/b/c/d/e/a/b/c/d/e"

	r2 := NewRouterV2()

	// var printed bool
	for _, r := range routes {
		r2.GET(r, func(c *Context) ResponseSender {
			// if !printed {
			// 	fmt.Println("match")
			// 	printed = true
			// }
			return nil
		})
	}

	h := NewHandler(r2)

	req, _ := http.NewRequest("GET", test, nil)
	resp := httptest.NewRecorder()

	b.StartTimer()
	for n := 0; n < b.N; n++ {
		h.ServeHTTP(resp, req)
	}
}
