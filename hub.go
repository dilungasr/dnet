package dnet

import "fmt"

// Hub is for maintaining chat app users
type Hub struct {
	contexts   map[*Context]bool
	rooms      map[string][]*Context
	register   chan *Context
	unregister chan *Context
	// checks to see if user has started dnet hub monitoring or not
	hasInitialized bool
}

// hub is a Hub instance to be used throught the application
var hub *Hub = &Hub{
	contexts:       make(map[*Context]bool),
	rooms:          make(map[string][]*Context),
	register:       make(chan *Context),
	unregister:     make(chan *Context),
	hasInitialized: false,
}

// Init initializes dnet hub monitoring. You should pass your configurations in this function
// if you are using v1.0.208-beta and above. Otherwise you should stick with Config() function
//
// It should be called only once
func Init(options ...Options) {
	// makesure that Init is called only once
	if hub.hasInitialized {
		panic("Dnet: Dnet cannot be initialized more than once")
	}

	//take ticketAge configurations if user has configured the time for ticket to expire
	if options[0].TicketAge > 0 {
		Router1.ticketAge = options[0].TicketAge
	}

	//take maximum message size configuration if user set
	if options[0].MaxSize > 0 {
		Router1.maxSize = options[0].MaxSize
	}

	//mark that Init has been called to prevent future repeated calling of this function
	hub.hasInitialized = true
	fmt.Println("Dnet initialized...")
	go hub.Run()
}

// Run method is for starting the Hub
func (h *Hub) Run() {
	for {
		select {
		case context := <-h.register:

			h.contexts[context] = true
		case context := <-h.unregister:
			if _, ok := h.contexts[context]; ok {
				delete(h.contexts, context)

				// close the send channel
				close(context.send)
			}

			// check if user is present in the rooms too ... unregister
			for room, contexts := range h.rooms {
				for index, ctx := range contexts {
					if ctx == context {
						// remove the context from the rooms
						if index != len(contexts)-1 {
							h.rooms[room] = append(contexts[:index], contexts[index+1:]...)
						} else {
							// if the context is the last one in the list  ...just slice it out
							h.rooms[room] = contexts[:index]

							//  check if after removing it the room became empty
							if len(h.rooms[room]) == 0 {
								// remove the room
								delete(h.rooms, room)
								close(context.send)
							}
						}
					}
				}
			}
		}
	}
}
