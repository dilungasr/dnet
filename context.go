package dnet

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Context is a middleman between the websocket connection and the Hub
type Context struct {
	// is for main	taining connections and rooms
	hub *Hub
	// send is for listening data sending to the Context
	send chan interface{}
	// values is for installing request values
	values map[string]interface{}
	//conn is a websocket connection
	conn *websocket.Conn

	// Action is the current action received from the client
	action string
	//ID is  a user id to assocaite it with the user connection
	ID string
	// Data is where the data sent from the client stored
	Data interface{}

	// Rec is the id of the recipient
	Rec string
	// next  is for telling to go to the next handler or not (in middlwares )
	goNext bool
	// authed is for telling if the user is authenticated or not
	authed bool
	// IP is the ip address of the user
	IP string

	// disposed tells wether the clinet context has been disposed or not
	disposed bool

	// loggedout tells wether the client context has been logged of or not
	loggedout bool
	// expireTime is the time the request expire if not authenticated
	expireTime time.Time
}

// Response is the structure  for the returned data
type Response struct {
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

// Message is a data sent from the client
type Message struct {
	Action string `json:"action"`
	// Rec is the id of the recipient
	Rec string `json:"rec"`
	// Data is the main payload to send to the recipient
	Data interface{} `json:"data"`
}

// readPump for reading the message from the websocket connection
func (c *Context) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// configure the connection values
	c.conn.SetReadLimit(Router1.maxSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// here we go ..... reading
	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Dnet: %v", err)

				// call the last seeen handler to update any last seen info
			}

			// logged out if authenticated,  where the client side connection has misbehaved
			if !c.disposed && !c.loggedout && c.authed {
				c.Logout()
			} else if !c.disposed && !c.loggedout && !c.authed {
				// dispose if  not authenticated,  where the client side connection has misbehaved
				c.Dispose()
			}

			break
		}

		// initialize and pour out the value from the dnet message to the context to make it available in the api context
		c.values = make(map[string]interface{})
		c.action = msg.Action
		c.Data = msg.Data
		c.Rec = msg.Rec

		// routing user action
		Router1.Route(msg.Action, c)
	}
}

// writePump for writing to the Context
func (c *Context) writePump() {
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
				log.Printf("Dnet error: %v", err)
				return
			}
		}

	}
}

// expireContet is for removing the expired context
func (c *Context) expireContext() {
	ticker := time.NewTicker(Router1.ticketAge)

	// remove the context if expired
	select {
	case <-ticker.C:
		if !c.authed {
			c.Dispose()
			ticker.Stop()
		}
	}
}

// Connect inits Dilunga Net's connection in the given endpoint
func Connect(w http.ResponseWriter, r *http.Request, allowedOrigin ...string) {

	//hub not started monitoring
	if !hub.hasInitialized {
		panic("Dnet: Dnet has not been initialized. Initialized dnet by calling the dnet.Init()")
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

	// create the Context...  mark user as not authenticated
	expireTime := time.Now().Local().Add(Router1.ticketAge)
	context := &Context{hub: hub, send: make(chan interface{}, 256), conn: conn, authed: false, expireTime: expireTime, disposed: false, loggedout: false}
	context.hub.register <- context

	go context.readPump()
	go context.writePump()
	go context.expireContext()
}
