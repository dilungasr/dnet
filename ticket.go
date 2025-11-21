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
	ExpireTime time.Time
}

// newRouterTicket creates an instance of routerTicket struct
func newRouterTicket(r *http.Request, ID string) (ticket routerTicket, err error) {
	UUID, err := uuid.NewRandom()
	if err != nil {
		return ticket, err
	}
	UUIDString := UUID.String()

	// create cipher text
	IP, err := GetIP(r)
	if err != nil {
		return ticket, err
	}

	t := time.Now().Add(router.ticketAge)
	expireTimeBytes, err := t.MarshalText()
	if err != nil {
		return ticket, err
	}

	expireTimeString := string(expireTimeBytes)

	ticketString := strings.Join([]string{ID, UUID.String(), IP, expireTimeString}, ",")

	cipherText, err := radi.Encrypt(ticketString, router.ticketSecrete)
	if err != nil {
		return ticket, err
	}

	//   save the router ticket in the router
	ticket = routerTicket{UUIDString, cipherText, t}
	router.tickets = append(router.tickets, ticket)

	return ticket, err
}

// hasExpired tells whether the given ticket time has expired or not
func hasExpired(t time.Time) bool {
	return time.Now().Local().After(t)
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
	devLogger("Ticket sent", ticket)

}

func devLogger(v ...any) {
	logs := []any{"[dnet-dev]"}
	logs = append(logs, v...)

	// print the stack trace
	// log.Println(string(debug.Stack()))
	log.Println(logs...)
}

// NewTicket creates and returns encrypted ticket
func NewTicket(r *http.Request, ID string) (ticket string, err error) {
	// create a new ticket
	rTicket, err := newRouterTicket(r, ID)
	if err != nil {
		return ticket, err
	}

	ticket = rTicket.CipherText
	devLogger("NewTicket created:", ticket)
	return ticket, err
}

// ticketParts splits the ticket string to get the indiviadial elements
func ticketParts(ticketString string, c ...*Ctx) (ID, UUID, IP string, expireTime time.Time, ok bool) {
	ticketPartsSlice := strings.Split(ticketString, ",")

	// if it is a false ticket with less or more number of elements of slice
	if len(ticketPartsSlice) != 4 {
		if len(c) > 0 {
			ctx := c[0]
			ctx.conn.SetWriteDeadline(time.Now().Local().Add(writeWait))
			ctx.SendBack(400, "Bad Request")
		}

		//output error to the console
		log.Println("[dnet] ticket string has unusual number of parts. It's invalid")
		return ID, UUID, IP, expireTime, false
	}

	//    organize and parse time for returning to the caller
	ID = ticketPartsSlice[0]
	UUID = ticketPartsSlice[1]
	IP = ticketPartsSlice[2]
	expireTime, err := time.Parse(time.RFC3339, ticketPartsSlice[3])
	if err != nil {
		if len(c) > 0 {
			ctx := c[0]
			ctx.conn.SetWriteDeadline(time.Now().Local().Add(writeWait))
			ctx.SendBack(400, "Ooop! Bad Request")
		}

		// log the error to the console
		log.Println("[dnet] ", err)
		return ID, UUID, IP, expireTime, false
	}

	return ID, UUID, IP, expireTime, true
}

// authenticateTicket authenticates the given encrypted ticket link from the client and returns userID and valid boolean
func authenticateTicket(c *Ctx, encryptedTicketString string) (userID string, valid bool) {
	devLogger("authenticateTicket called")
	//get the ticket string from the client to plain text
	ticketString, err := radi.Decrypt(encryptedTicketString, router.ticketSecrete)
	// if the ticketString is not avalid base64 string
	if err != nil {
		c.SendBack(400, "Bad Request")
		return userID, false
	}
	devLogger("authenticateTicket ticketString:", ticketString)

	// split the ticket string to individal parts
	userID, clientUUID, IP, expireTime, ok := ticketParts(ticketString)
	devLogger("UserID:", userID, "clientUUID:", clientUUID, "IP:", IP, "expireTime:", expireTime, "ok:", ok)
	if !ok || IP != c.IP {
		return userID, false
	}

	ticket, found := router.findTicket(clientUUID, encryptedTicketString)
	devLogger("Is ticket found:", found)
	if !found {
		return userID, false
	}

	//  check if not expired
	if hasExpired(expireTime) {
		devLogger("Ticket expired")
		//  remove the ticket
		router.removeTicket(ticket.UUID, ticket.CipherText)
		return userID, false
	}

	//remove the ticket
	router.removeTicket(ticket.UUID, ticket.CipherText)
	// mark authed
	c.ID = userID
	c.Authed = true

	devLogger("Ticket authenticated...")

	// return success to the caller
	return userID, true
}
