package dnet

import (
	"fmt"
	"log"

	"github.com/mitchellh/mapstructure"
)

// By Dilunga SR<dilungasr@gmail.com>
// wwww.axismedium.com
// twitter: @dilungasr

// Broadcast sends data to all execept the sender
func (c *Ctx) Broadcast(statusAndData ...interface{}) {
	res := prepareRes(c, "Broadcast", statusAndData)

	// pass to all hub contexts to send to all other contexts
	for context := range c.hub.contexts {
		// send to other contexts except this
		if context != c {
			select {
			case context.send <- res:
			default:
				deleteContext(context)
			}
		}
	}
}

// All sends to anyone connected to the websocket Dnet instance including the sender of the message
func (c *Ctx) All(statusAndData ...interface{}) {
	res := prepareRes(c, "All", statusAndData)
	// pass to all hub contexts to send to all other contexts
	for context := range c.hub.contexts {
		select {
		case context.send <- res.checkSource(c, context):
		default:
			deleteContext(context)
		}
	}
}

// SendBack sends back to the sender's sending connection only.
// It sends to the original action of the sender. Use SendSelf() if you want to send to the current action.
func (c *Ctx) SendBack(statusAndData ...interface{}) {
	res := prepareRes(c, "SendBack", statusAndData, true)
	res.setSource(c)

	// send back to the client

	select {
	case c.send <- res:
	default:
		go func() {
			c.send <- res
		}()
	}
}

// SendSelf sends to the current context's action. Similar to SendBack() but it does
// not send to the original action but to the current action.
func (c *Ctx) SendSelf(statusAndData ...interface{}) {
	res := prepareRes(c, "SendSelf", statusAndData)
	res.setSource(c)

	// send back to the client

	select {
	case c.send <- res:
	default:
		go func() {
			c.send <- res
		}()
	}
}

// Send sends to one client only
func (c *Ctx) Send(ID string, statusAndData ...interface{}) {
	res := prepareRes(c, "Send", statusAndData)

	// find the user to which the dataIndex should be sent to
	for context := range c.hub.contexts {
		if context.ID == ID {
			select {
			case context.send <- res.checkSource(c, context):
			default:
				deleteContext(context)
			}
		}
	}

}

// FilterFunc returns true to include or false to filter out
type FilterFunc func(c *Ctx) bool

// SendByFilter sends to every context where the filter returns true
func (c *Ctx) SendByFilter(filter FilterFunc, statusAndData ...interface{}) {
	res := prepareRes(c, "SendByFilter", statusAndData)

	// find the user to which the dataIndex should be sent to
	for context := range c.hub.contexts {
		if filter(context.CloneWithEmptyValues()) {
			select {
			case context.send <- res.checkSource(c, context):
			default:
				deleteContext(context)
			}
		}
	}

}

// Calls senderFunc for each context on the dnet hub and passes it on the function
func (c *Ctx) SendByFunc(senderFunc ActionHandler) {
	for context := range c.hub.contexts {
		senderFunc(context.CloneWithEmptyValues())
	}
}

// SendMe sends to all of the sender's open connections
func (c *Ctx) SendMe(statusAndData ...interface{}) {
	// send to the sender's connection first
	c.SendSelf(statusAndData...)

	// send to all other connections except the sender
	res := prepareRes(c, "SendMe", statusAndData)

	// find the user to which the dataIndex should be sent to
	for context := range c.hub.contexts {
		if context.ID == c.ID && context != c {
			select {
			case context.send <- res:
			default:
				deleteContext(context)
			}
		}
	}

}

// Multicast sends to the given users IDs (useful for sharing something to multiple users
func (c *Ctx) Multicast(userIDs []string, statusAndData ...interface{}) {
	res := prepareRes(c, "Multicast", statusAndData)

	for _, userID := range userIDs {
		// find the matching context
		for context := range c.hub.contexts {
			if userID == context.ID {
				select {
				case context.send <- res.checkSource(c, context):
				default:
					deleteContext(context)
				}
			}
		}
	}
}

/*
   -----------------------------------------------
    ROOM  METHODS GOES HERE
   -----------------------------------------------

*/

// RoomAll sends to the members of the room. (useful for chat rooms.. and sending data to all people under the same role or cartegory )
func (c *Ctx) RoomAll(ID string, statusAndData ...interface{}) {
	res := prepareRes(c, "RoomAll", statusAndData)

	//    find the room and broadcast to all the room members
	for roomID, contexts := range c.hub.rooms {
		if roomID == ID {
			//   broadcast to all the members in the room
			for _, context := range contexts {
				select {
				case context.send <- res.checkSource(c, context):
				default:
					deleteContext(context)
				}
			}

			// break out of the loop if found the room and finished sending to all members of the room
			break
		}
	}
}

// RoomBroadcast sends to all members of the registered room except the sender
func (c *Ctx) RoomBroadcast(ID string, statusAndData ...interface{}) {
	res := prepareRes(c, "Send", statusAndData)

	//    find the room and broadcast to all the room members
	for roomID, contexts := range c.hub.rooms {
		if roomID == ID {
			//   broadcast to all the members in the room
			for _, context := range contexts {
				// send to all members of the room execept the sender
				if context != c {
					select {
					case context.send <- res.checkSource(c, context):
					default:
						deleteContext(context)
					}
				}
			}
		}
	}
}

// CreateRoom is for creating a new room.... if it finds a room exist it only adds the given the room
func (c *Ctx) CreateRoom(roomID string, usersIDS ...string) {
	isReg := false

	// if the room is already registered
	for room, contexts := range c.hub.rooms {
		if room == roomID {
			isReg = true
			// find active user contexts to add to the room
			for context := range c.hub.contexts {
				for _, id := range usersIDS {
					if id == context.ID {

						// do the room found already added in the room?
						found := false

						//check if the context already added in the room
						for _, roomContext := range contexts {
							if roomContext.ID == id {
								found = true
								break
							}
						}

						// add to the room if hasn't already
						if !found {
							contexts = append(contexts, context)
						}
						break
					}
				}
			}
		}
	}

	// do not the code below if the room exist
	if isReg {
		return
	}

	// if the room not found .... create a new room
	activeUsers := []*Ctx{}
	for context := range c.hub.contexts {
		for _, id := range usersIDS {
			// if finds an active user
			if id == context.ID {
				activeUsers = append(activeUsers, context)
				break
			}
		}
	}

	// create an active room only when there are active room members
	if len(activeUsers) > 0 {
		c.hub.rooms[roomID] = activeUsers
	}
}

// Next pushes the next middleware in the list
func (c *Ctx) Next() {
	c.goNext = true
}

// Rooms assigns this connnection to the chatrooms it relates to
func (c *Ctx) Rooms(roomsIDs ...string) {

	// if user has rooms
	if len(roomsIDs) > 0 {

		for _, room := range roomsIDs {
			// find if already added and only append user if found
			isReg := false
			for registeredRoom, contexts := range c.hub.rooms {
				if registeredRoom == room {
					isReg = true

					//  check if the context already added to the room
					found := false
					for _, context := range contexts {
						if context == c {
							found = true
						}
					}

					// add the context to the room if no registered
					if !found {
						// append user to the rooms
						c.hub.rooms[registeredRoom] = append(contexts, c)

					}
					break
				}
			}

			//    if chat room is added for the first time in the hub
			if !isReg {
				c.hub.rooms[room] = []*Ctx{c}
			}
		}
	}
}

/*
   -----------------------------------------------
   TICKET METHODS GEOS HERE
   -----------------------------------------------

*/

// ticket is for converting the received golang interface{} data to ticket ..... get the fields
type clientTicket struct {
	Ticket string `json:"ticket"`
}

// AuthTicket authenticates the ticket and  returns the user ID provided in the SendTicket().
//
// ok is also returned to inform you if user authentication fails or not. If failed you, all you can do
// is simply "return" to finish your reponse cycle since AuthTicket helps you to respond to the client
// with the infoText you provide or the default infoText
//
// Infotext is a message to send back to the client when user authentication fails.
// It defaults to "Please login to access this resource" if you do not provide it.
func (c *Ctx) AuthTicket(infoText ...string) (ID string, ok bool) {
	var ticketFromClient clientTicket

	if len(infoText) == 0 {
		infoText = []string{"Please login to access this resource"}
	}

	if !c.Binder(&ticketFromClient) {
		return
	}

	// auth user ticket
	ID, ok = authenticateTicket(c, ticketFromClient.Ticket)
	if ok {
		return ID, ok
	}

	// if not valid ... close the connection
	c.SendBack(401, infoText[0])
	c.Dispose()
	return "", false
}

// MarkAuthed marks this connection as authenticated. Hence, no ticket authentication required.
//
// Use this when you have your own authentication mechanism or just for testing purposes.
func (c *Ctx) MarkAuthed(ID string) {
	c.ID = ID

	c.Authed = true
}

/*
   -----------------------------------------------
    FIRE METHODS GOE HERE
   -----------------------------------------------

*/

// Fire sets which action to fire to the client. It's recommended to keep the action in path form to maintain maintain uniformity,
// If you do not set the action to fire, the action you listened for it will be fired backward to the client too.
func (c *Ctx) Fire(action string) {
	c.action = action
}

// Refire resets action to the initial action before calling any Fire("/action") method.
// In fact, the action becomes the same as it was fired by the client.
func (c *Ctx) Refire() {
	c.action = c.GetOriginalAction()
}

/*
   -----------------------------------------------
    BINDING METHODS GOE HERE
   -----------------------------------------------

*/

// Binder extracts raw data, stores it in the passed pointer v
// and implicitly sends 422 Unprocessable entities to the client when it's unable to decode data. You can provide a custom
// client error message by passing it as msg argument. In addition to sending a client friendly error to the client it also logs
// a raised error in console
//
// It returns true on success and false on failure. You can use this value to stop the execution in
// the caller function by just using a return statement.
//
// Use Bind(v interface{}) if you want to explicitly handle errors
func (c *Ctx) Binder(v interface{}, msg ...string) (ok bool) {
	if len(msg) == 0 {
		msg = []string{"Unprocessable entities"}
	}

	if err := mapstructure.Decode(c.data, v); err != nil {
		log.Println("[dnet] ", err)
		c.SendBack(422, msg)
		return false
	}

	return true
}

// Bind extracts raw data, stores it in the passed pointer v
// and returns any raised error
func (c *Ctx) Bind(v interface{}) (err error) {
	return mapstructure.Decode(c.data, v)
}

/*
   -----------------------------------------------
    CLOSE METHOD GOES HERE
   -----------------------------------------------

*/

// Dispose discards the client connection without calling LastSeen for saving any lastSeen info for the clinet connection
// Useful for expired unauthorized client connections
func (c *Ctx) Dispose() {
	c.disposed = true
	c.hub.unregister <- c
	c.conn.Close()
}

// Logout calls the LastSeen function to ensure user last seen data is saved before discarding the client connection
func (c *Ctx) Logout() {
	c.loggedout = true
	// set the last seen of the clinet connection
	router.lastSeen(c)

	// unregister the clinet context
	c.hub.unregister <- c
	c.conn.Close()
}

/*
  -----------------------------------------------------------
  |  SETTING VALUES IN AND GETTING VALUES FROM CONNECTION   |
  ----------------------------------------------------------
*/

// Set stores value in the connection.
func (c *Ctx) Set(key string, val interface{}) {
	c.values[key] = val
}

// Get gets data stored in the connection
func (c *Ctx) Get(key string) (val interface{}, err error) {
	val, ok := c.values[key]
	if !ok {
		return val, fmt.Errorf("[dnet] " + "'" + key + "'" + " field is not registered in the context")
	}

	return val, nil

}

// Del deletes context value's field with a given keyc
func (c *Ctx) Del(key string) {
	delete(c.values, key)
}

// Clear empties context values. That is, deletes every value in the context values
// and resets the value store anew.
func (c *Ctx) Clear() {
	c.values = map[string]interface{}{}
}

/*
  -----------------------------------------------------------
  |  WORKING WITH EMAILS   |
  ----------------------------------------------------------
*/
