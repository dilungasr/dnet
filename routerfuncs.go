package dnet

import (
	"net/http"
	"time"

	"github.com/dilungasr/radi"
)

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

// SendTicket sends an encrypted ticket to the user.
func SendTicket(r *http.Request, w http.ResponseWriter, ID string) {
	// if the ticketSecrete and the iv set...... wer are ready to go
	secreteKey := router.ticketSecrete
	iv := router.ticketIV

	// plain data(not encrypted)
	expireTimeBytes, err := time.Now().Local().Add(router.ticketAge).MarshalText()
	if err != nil {
		panic(err)
	}
	expireTime := string(expireTimeBytes)
	// get the user IP
	IP, err := GetIP(r)
	if err != nil {
		jsonSender(w, 400, Msg("Bad Request!"))
		return
	}

	// encrypt the tikcet data to be sent to the client
	newTicket := ID + "," + IP + "," + expireTime

	encTicket := radi.Encrypt(newTicket, secreteKey, iv)

	// save the ticket in the router
	router.tickets = append(router.tickets, newTicket)

	// send the ticket to the client
	jsonSender(w, 200, Map{"ticket": encTicket})

}
