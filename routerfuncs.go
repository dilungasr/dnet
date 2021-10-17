package dnet

import (
	"net/http"
	"time"

	"github.com/dilungasr/radi"
)

// On method is adding the Event handlers to the router
func On(action string, handlers ...ActionHandler) {
	Router1.actionHandlers[action] = handlers
}

// Router is for grouping the actions by matching their paths
func Router(path string) RouterMatcher {
	return RouterMatcher{path}
}

// Use is for adding middlewares to the root of the dnet action path
func Use(handlers ...ActionHandler) {
	Router1.routeMatchers["/"] = append(Router1.routeMatchers["/"], handlers...)
}

// SendTicket sends an encrypted ticket to the use and saves the clean one the router
func SendTicket(r *http.Request, w http.ResponseWriter, ID string) {
	// if the ticketSecrete and the iv set...... wer are ready to go
	secreteKey := Router1.ticketSecrete
	iv := Router1.ticketIV

	// plain data(not encrypted)
	expireTimeBytes, err := time.Now().Local().Add(Router1.ticketAge).MarshalText()
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

	// save the ticket in the Router1
	Router1.tickets = append(Router1.tickets, newTicket)

	// send the ticket to the client
	jsonSender(w, 200, Map{"ticket": encTicket})

}
