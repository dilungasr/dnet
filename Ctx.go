package dnet

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// By Dilunga SR<dilungasr@gmail.com>
// wwww.axismedium.com
// twitter: @dilungasr

// Ctx is a middleman between the websocket connection and the Hub.
// Ctx is stored in the dnet hub and hence it is an inside and persistent context.
type Ctx struct {
	// is for main	taining connections and rooms
	hub *Hub
	// send listens data coming to the context from the hub
	send chan interface{}
	// values is for installing request values
	values map[string]interface{}
	//conn is a websocket connection
	conn *websocket.Conn

	// Action is the action to fire
	action string
	//ID is  a user id to assocaite with the user connection
	ID string
	// Data stores data received from the client side
	Data interface{}

	// Rec is an id of the recipient
	Rec string
	// goNext  tells whether to go to the next handler or not (in middlwares )
	goNext bool
	// Authed  tells if the connection is authenticated or not
	Authed bool
	// IP is the ip address of the user
	IP string

	// disposed tells wether the client context has been disposed or not
	disposed bool

	// loggedout tells wether the client context has been logged of or not
	loggedout bool
	// expireTime is the time for request to expire if not authenticated yet
	expireTime time.Time
}

// response models data sent to the client.
type response struct {
	// Action is
	Action string `json:"action"`
	//the status  code to be returned to the client
	Status int `json:"status"`
	// the true payload to be sent to the client
	Data interface{} `json:"data"`
	// Sender is the id of the sender
	Sender string `json:"sender"`
}

// constants
const (
	// writeWait is the maximum time to wait writing to the peer
	writeWait = 10 * time.Second
	// pongWait muximum time to wait for the pongMessage from the peer
	pongWait = 60 * time.Second
	//pingWait is the time to wait before sending the next pingMessage to the peer... Must be smaller than the pongWait
	pingPeriod = (9 * pongWait) / 10
)

var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
}

// message is a data sent from the client
type message struct {
	Action string `json:"action"`
	// Rec is the id of the recipient
	Rec string `json:"rec"`
	// Data is the main payload to send to the recipient
	Data interface{} `json:"data"`
}

// readPump for reading the message from the websocket connection
func (c *Ctx) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// configure the connection values
	c.conn.SetReadLimit(router.maxSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// here we go ..... reading
	for {
		var msg message
		err := c.conn.ReadJSON(&msg)
		if err != nil {

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("dnet: %v", err)

				// call the last seeen handler to update any last seen info
			}

			// logged out if authenticated,  where the client side connection has misbehaved
			if !c.disposed && !c.loggedout && c.Authed {
				c.Logout()
			} else if !c.disposed && !c.loggedout && !c.Authed {
				// dispose if  not authenticated,  where the client side connection has misbehaved
				c.Dispose()
			}

			break
		}

		// initialize and pour out the value from the dnet message to the context to make it available in the api context
		if c.values == nil {
			c.values = make(map[string]interface{})
		}
		c.action = msg.Action
		c.Data = msg.Data
		c.Rec = msg.Rec

		// routing user action
		router.Route(msg.Action, c)
	}
}

// writePump for writing to the Ctx
func (c *Ctx) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteJSON(message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// write the ping message
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("dnet error: %v", err)
				return
			}
		}

	}
}

// expireContext disposes this context when tickage age reaches without being authenticated
func (c *Ctx) expireContext() {
	<-time.After(router.ticketAge)

	if !c.Authed {
		c.Dispose()
	}
}

// Connect inits Dilunga Net's connection in the given endpoint
func Connect(w http.ResponseWriter, r *http.Request, allowedOrigin ...string) {

	//hub not started monitoring
	if !hub.hasInitialized {
		panic("dnet: dnet has not been initialized. Initialized dnet by calling the dnet.Init()")
	}

	// PROTECT UNAUTHORIZED ORIGINS
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// if no origin allowed  ...cancel any connection
		if len(allowedOrigin) < 1 {
			return false
		}

		// if there are allowed origins ... match the origin with the incomiing one
		for _, origin := range allowedOrigin {
			if origin == r.Host {
				return true
			}
		}

		// if the host is not allowed
		return false
	}

	// upgrade the http connnection to websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// create the Ctx...  mark user as not authenticated
	expireTime := time.Now().Local().Add(router.ticketAge)
	context := &Ctx{hub: hub, send: make(chan interface{}, 256), conn: conn, Authed: false, expireTime: expireTime, disposed: false, loggedout: false}
	context.hub.register <- context

	go context.readPump()
	go context.writePump()
	go context.expireContext()
}
