package dnet

//On method takes ActionHandlers to be called when the given action fired by the dnet-client
func On(action string, handlers ...ActionHandler) {
	router.actionHandlers[action] = handlers
}

// Router creates a subrouter for grouping related actions.
func Router(path string) Subrouter {
	return Subrouter{path}
}

// Use adds root-level middlewares which will be called before any action is matched.
func Use(handlers ...ActionHandler) {
	router.routeMatchers["/"] = append(router.routeMatchers["/"], handlers...)
}
