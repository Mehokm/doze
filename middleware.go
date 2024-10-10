package doze

// MiddlewareFunc defines a function to process middleware in the request chain.
type MiddlewareFunc func(ctx *Context, next NextFunc)

// NextFunc defines a function to proceed to the next middleware in the chain.
type NextFunc func(ctx *Context)

// ActionFunc defines the final action function to handle the request.
type ActionFunc func(ctx *Context) ResponseSender

// middlewareChain manages a chain of middlewares and the final action.
type middlewareChain struct {
	middlewares []MiddlewareFunc
	action      ActionFunc
}

// use adds a MiddlewareFunc to the middleware chain.
func (mc *middlewareChain) use(middleware MiddlewareFunc) {
	mc.middlewares = append(mc.middlewares, middleware)
}

// run executes the middleware chain and the final action.
func (mc *middlewareChain) run(ctx *Context) {
	// Start with the final action as the initial next function.
	next := func(ctx *Context) {
		executeAction(ctx, mc.action)
	}

	// Build the middleware chain in reverse order.
	for i := len(mc.middlewares) - 1; i >= 0; i-- {
		m := mc.middlewares[i]
		currentMiddleware := m
		currentNext := next
		next = func(ctx *Context) {
			currentMiddleware(ctx, currentNext)
		}
	}

	// Execute the middleware chain.
	next(ctx)
}

// executeAction calls the final action function and sends the response.
func executeAction(ctx *Context, action ActionFunc) {
	result := action(ctx)
	if result != nil {
		if _, err := result.Send(ctx.ResponseWriter); err != nil {
			panic(err)
		}
	}
}
