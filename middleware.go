package doze

// MiddlewareFunc is a type for adding middleware for the request
type MiddlewareFunc func(*Context, NextFunc)
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
	fn := buildChain(ctx, mc.root, mc.action)

	// start the chain
	fn(ctx)
}

func buildChain(ctx *Context, m *middleware, action ActionFunc) NextFunc {
	return func(ctx *Context) {
		if m.next == nil {
			// no more middleware, so call the actual action
			m.fn(ctx, func(ctx *Context) {
				result := action(ctx)

				if result != nil {
					_, err := result.Send(ctx.ResponseWriter)

					if err != nil {
						panic(err)
					}
				}
			})
		} else {
			// keep building the chain...
			m.fn(ctx, buildChain(ctx, m.next, action))
		}
	}
}
