package dnet

import (
	"regexp"
	"time"

	"github.com/dilungasr/radi"
)

//ActionHandler is a function wich receives Ctx
type ActionHandler func(*Ctx)

// Options is used to take all the
type Options struct {
	TicketAge time.Duration
	MaxSize   int64
}

// MainRouter routes websocket actions
type MainRouter struct {
	// routeMatchers is matching the middleware the given path prefixes
	routeMatchers map[string][]ActionHandler

	// actionHandlers are handlers to executed for the particular action
	actionHandlers map[string][]ActionHandler
	// lastSeenHandler is called when the connection gets closed
	lastSeenHandler ActionHandler
	// tells wether use set the last seen or not
	isLastSeenHandlerSet bool
	// tickets is the store place of all tickets given to the user before they expire
	tickets []routerTicket
	// ticketKey is the key for cryptography of the ticket
	ticketIV []byte

	// max message size
	maxSize int64

	// ticketSecrete is the secrete key for tecket encryption
	ticketSecrete string
	// ticketAge is the time the ticket expires
	ticketAge time.Duration
}

//Route performs routing websocket actions based on the incoming action
func (r *MainRouter) Route(IncomingAction string, context *Ctx) {
	if len(r.routeMatchers) > 0 {
		for path, handlers := range r.routeMatchers {

			// the incoming action must start with the router matcher's path
			regex := regexp.MustCompile("^" + path)

			//check if the incoming action matches the path in the beginning
			if regex.MatchString(IncomingAction) {
				for _, handler := range handlers {
					// set the goNext... c.Next() should be called to change the goNext value to true for the middleware to move to the next one
					// Otherwise the middleware will not proceed to the next middlware.
					context.goNext = false
					// call the handler
					handler(context)

					// stop if the has not passed the middleware... ie c.Next() has not been called
					if !context.goNext {
						return
					}
				}
			}
		}
	}

	//  go on to  the normal event handlers
	// isMatch is for generating 404 error if the message didn't find any action
	isMatch := false
	for handlerAction, handlers := range r.actionHandlers {
		// check for route matching
		if handlerAction == IncomingAction {
			isMatch = true

			// if there are middlewares
			if len(handlers) > 1 {
				for _, handler := range handlers {
					// set the goNext... c.Next() should be called to change the goNext value to true for the middleware to move to the next one
					// Otherwise the middleware will not proceed to the next middlware.
					context.goNext = false
					// call the handler
					handler(context)

					// stop if the has not passed the middleware... ie c.Next() has not been called
					if !context.goNext {
						return
					}

				}
			} else if len(handlers) == 1 {

				//if the action has only one actionHandler
				handlers[0](context)

			} else {
				// if no handler by the user
				panic("Dnet: No action handler passed")
			}
		}
	}

	// if the action matched nothing ..... return the 404 code to the client
	if !isMatch {
		context.conn.WriteJSON(response{IncomingAction, 404, "Action Not Found", ""})
	}

}

// LastSeen is called when the connection gets closed.
// It's very useful for setting the last seen or apperance of the user
func (r *MainRouter) lastSeen(c *Ctx) {
	// don't call last seen for non-authed connection
	if !c.Authed {
		return
	}
	// check to see if there are more than  1 context in the hub
	contextCount := 0

	for context := range c.hub.contexts {
		if context.ID == c.ID {
			contextCount++
		}
	}

	//if there are no any other contexts online ...then call the last seeen handler
	if contextCount == 1 {
		//check if lastSeen hanlder is set
		if r.isLastSeenHandlerSet {
			r.lastSeenHandler(c)
		}
	}
}

// LastSeen is called when the authenticated user gets offline or logs out
// It's very useful for setting the last seen of the user connection
func LastSeen(handler ActionHandler) {
	//  the lastSeenHandler is set
	router.isLastSeenHandlerSet = true

	// set it
	router.lastSeenHandler = handler
}

// ticketCleaner cleans expired tickets
func (r *MainRouter) ticketCleaner() {
	ticker := time.NewTicker(r.ticketAge)

	for range ticker.C {

		for _, ticket := range r.tickets {
			// remove if expired
			if hasExpired(ticket.ExpireTime) {
				router.removeTicket(ticket.UUID, ticket.CipherText)
			}
		}

	}

}

// TICKET'S ROUTER METHODS

// findTicket queries a ticket store for the ticket and returns it if found
//
// found - tells whehter the query found any match or not
func (router MainRouter) findTicket(UUID, cipherText string) (foundTicket routerTicket, found bool) {
	tickets := router.tickets

	for _, ticket := range tickets {
		if ticket.UUID == UUID && ticket.CipherText == cipherText {
			return ticket, true
		}
	}

	return foundTicket, false
}

// findTicketIndex queries a ticket store for the ticket and returns it's index and tells  it if found or not
//
// found - tells whehter the query found any match or not
func (router MainRouter) findTicketIndex(UUID, cipherText string) (index int, found bool) {
	tickets := router.tickets

	for i, ticket := range tickets {
		if ticket.UUID == UUID && ticket.CipherText == cipherText {
			return i, true
		}
	}

	return index, false
}

// removeTicket removes the ticket with the given uuid and cipher text from the store.
//It does not retain the order of the tickets
func (router *MainRouter) removeTicket(UUID, cipherText string) {
	// find the ticket index
	index, found := router.findTicketIndex(UUID, cipherText)
	if found {
		//remove the element at the index
		tickets := router.tickets
		tickets[index] = tickets[len(tickets)-1]
		router.tickets = tickets[:len(tickets)-1]
	}

}

// app-wise router
var router *MainRouter = &MainRouter{
	actionHandlers: make(map[string][]ActionHandler), routeMatchers: make(map[string][]ActionHandler),
	ticketSecrete: radi.RandString(32), ticketIV: radi.RandBytes(16),
	ticketAge: 30 * time.Second,
	maxSize:   512,
}
