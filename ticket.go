package dnet

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dilungasr/radi"
	"github.com/google/uuid"
)

// routerTicket holds the uuid of the ticket and the encrypted ticket string
type routerTicket struct {
	UUID       string
	CipherText string
}

// newRouterTicket creates an instance of routerTicket struct
func newRouterTicket(r *http.Request, ID string) (ticket routerTicket, err error) {
	UUID, err := uuid.NewRandom()
	UUIDString := UUID.String()

	// create cipher text
	IP, err := GetIP(r)
	expireTimeBytes, err := time.Now().Add(router.ticketAge).MarshalText()
	expireTimeString := string(expireTimeBytes)
	ticketString := strings.Join([]string{ID, UUID.String(), IP, expireTimeString}, ",")

	cipherText := radi.Encrypt(ticketString, router.ticketSecrete, router.ticketIV)

	//   save the router ticket in the router
	ticket = routerTicket{UUIDString, cipherText}
	router.tickets = append(router.tickets, ticket)

	return ticket, err
}

// SendTicket sends an encrypted ticket to the user.
func SendTicket(r *http.Request, w http.ResponseWriter, ID string) {
	// create a new ticket
	ticket, err := newRouterTicket(r, ID)
	if err != nil {
		log.Println(err)
		jsonSender(w, 500, "Unable to complete authentication operation. Please try again later.")
		return
	}

	// send the ticket to the client
	jsonSender(w, 200, Map{"ticket": ticket.CipherText})

}
