package dnet

// Subrouter is for grouping actions, and creating middleware subjected to the particular group
type Subrouter struct {
	prefix string
}

func (r *Subrouter) Router(path string) Subrouter {
	fullAction := r.prefix + path
	return Subrouter{fullAction}
}

// On method is adding Event handlders to the router by prefixing it with the Matcher path
func (r Subrouter) On(action string, handlers ...ActionHandler) {
	router.actionHandlers[r.prefix+action] = handlers
}

//Use adds middlewares to the subrouter which will be called before any other subrouter actions
func (r *Subrouter) Use(handlers ...ActionHandler) {
	// append the given action handlers for matching the actioon-path beginning
	router.routeMatchers[r.prefix] = append(router.routeMatchers[r.prefix], handlers...)
}
