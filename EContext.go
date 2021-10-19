package dnet

// EContext is an external context which provides functionalities for sending data to the websocket's
// connections from a normal http connection. It interfaces the dnet hub and dnet sending functionalities to the normal http connection.
// EContext is not stored inside the dnet hub and hence it is called external and immidiate context.
type EContext struct {
	// is for main	taining connections and rooms
	hub *Hub
	// Action is the current action received from the client
	action string
	//ID is  a user id to assocaite it with the user connection
	ID string
}

// Context creates an external context for sending data to the websocket connections from a normal http connection.
func Context(action string, userID ...string) EContext {
	ID := ""
	if len(userID) > 0 {
		ID = userID[0]
	}

	return EContext{hub, action, ID}
}
