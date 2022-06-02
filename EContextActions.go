package dnet

// Written by Dilunga SR

// Broadcast sends data to alll dnet contexts in the hub
func (c *EContext) Broadcast(statusAndData ...interface{}) {
	res := prepareRes(c, "Broadcast", statusAndData)

	// pass to all hub contexts to send to all other contexts
	for context := range c.hub.contexts {
		// send to other contexts except this
		select {
		case context.send <- res:
		default:
			deleteContext(context)
		}

	}
}

// Send sends to only one client of the specified ID
func (c *EContext) Send(ID string, statusAndData ...interface{}) {
	res := prepareRes(c, "Send", statusAndData)

	// find the user to which the dataIndex should be sent to
	for context := range c.hub.contexts {
		if context.ID == ID {
			select {
			case context.send <- res:
			default:
				deleteContext(context)
			}
		}
	}

}

// Multicast sends to the given users IDs (useful for sharing something to multiple users
func (c *EContext) Multicast(userIDs []string, statusAndData ...interface{}) {
	res := prepareRes(c, "Multicast", statusAndData)

	for _, userID := range userIDs {
		// find the matching context
		for context := range c.hub.contexts {
			if userID == context.ID {
				select {
				case context.send <- res:
				default:
					deleteContext(context)
				}
			}
		}
	}
}

/*
   -----------------------------------------------
    FIRE METHODS GOE HERE
   -----------------------------------------------

*/

// Fire sets which action to fire to the client. It's recommended to keep the action in path form to maintain maintain uniformity,
// If you do not set the action to fire, the action you listened for it will be fired backward to the client too.
func (c *EContext) Fire(action string) {
	c.action = action
}

/*
  -----------------------------------------------------------
  |  WORKING WITH EMAILS   |
  ----------------------------------------------------------
*/

// By Dilunga SR
