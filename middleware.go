package doze

// MiddlewareFunc is a function type for adding middleware to the request
type MiddlewareFunc func(*Context, NextFunc)

// NextFunc is a function type that progresses the middleware chain for the request
type NextFunc func(*Context)

type middleware struct {
	fn   MiddlewareFunc
	next *middleware
}

type middlewareChain struct {
	root   *middleware
	action ActionFunc
}

func (mc *middlewareChain) add(mf MiddlewareFunc) {
	if mc.root == nil {
		mc.root = &middleware{
			fn: mf,
		}
	} else {
		tmp := mc.root
		for tmp.next != nil {
			tmp = tmp.next
		}

		tmp.next = &middleware{
			fn: mf,
		}
	}
}

func (mc *middlewareChain) run(ctx *Context) {
	// build the actual function chain
	mcCopy := *mc

	fn := buildChain(ctx, mcCopy.root, mcCopy.action)

	// start the chain
	fn(ctx)
}

func buildChain(ctx *Context, m *middleware, action ActionFunc) NextFunc {
	return func(ctx *Context) {
		// no middleware, so no need to wrap in MiddlewareFunc
		if m == nil {
			doAction(ctx, action)
		} else if m.next == nil {
			// no more middleware, so call the actual action
			m.fn(ctx, func(ctx *Context) {
				doAction(ctx, action)
			})
		} else {
			// keep building the chain recursively...
			m.fn(ctx, buildChain(ctx, m.next, action))
		}
	}
}

func doAction(ctx *Context, action ActionFunc) {
	result := action(ctx)

	if result != nil {
		_, err := result.Send(ctx.ResponseWriter)

		if err != nil {
			panic(err)
		}
	}
}
