package dnet

// EContext is an external context which provides functionalities for sending data to the websocket's
// connections from a normal http connection. It interfaces the dnet hub and dnet sending functionalities to the normal http connection.
// EContext is not stored inside the dnet hub and hence it is called external and immidiate context.
type EContext struct {
	// is for main	taining connections and rooms
	hub *Hub
	// Action is the current action received from the client
	action string
	//original action stays the same as it was passed in the Context()
	originalAction string
	//ID is  a user id to assocaite it with the user connection
	ID string
}

func (ec EContext) getAction() string {
	return ec.action
}
func (ec EContext) getOriginalAction() string {
	return ec.originalAction
}

func (ec EContext) getID() string {
	return ec.ID
}

func (ec EContext) getAsyncID() string {
	return ""
}

// Context creates an external context for sending data to the websocket connections from a normal http connection.
func Context(action string, userID ...string) EContext {
	ID := ""
	if len(userID) > 0 {
		ID = userID[0]
	}

	return EContext{
		hub:            hub,
		action:         action,
		originalAction: action,
		ID:             ID,
	}
}
