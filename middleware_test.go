package doze

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddlewareChain(t *testing.T) {
	// given
	response := httptest.NewRecorder()

	ctx := &Context{
		Request:        httptest.NewRequest("GET", "http://test", nil),
		ResponseWriter: &ResponseWriter{response, 0, 0},
	}

	mwc := &middlewareChain{
		action: stubActionFunc,
	}

	// when
	mwc.add(middlewareFuncOne)
	mwc.add(middlewareFuncTwo)
	mwc.add(middlewareFuncThree)

	mwc.run(ctx)

	// then

	assert.Equal(t, []byte(`{"Value":3}`), response.Body.Bytes(), "should be equal")
}

func middlewareFuncOne(ctx *Context, next NextFunc) {
	ctx.Set("one", 11)

	next(ctx)
}

func middlewareFuncTwo(ctx *Context, next NextFunc) {
	ctx.Set("two", ctx.Value("one").(int)-2)

	next(ctx)
}

func middlewareFuncThree(ctx *Context, next NextFunc) {
	ctx.Set("three", ctx.Value("two").(int)/3)

	next(ctx)
}

func stubActionFunc(ctx *Context) ResponseSender {
	result := struct {
		Value int
	}{
		ctx.Value("three").(int),
	}

	return NewOKJSONResponse(result)
}
