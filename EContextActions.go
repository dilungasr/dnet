package dnet

// Written by Dilunga SR

// Broadcast sends data to alll dnet contexts in the hub
func (c *EContext) Broadcast(statusAndData ...interface{}) {
	dataIndex := 0
	statusCode := 200

	// take user dataIndex from the statusAndCode and assign them to the above variables
	assignData(&dataIndex, &statusCode, statusAndData, "Broadcast")
	// prepare the response to be sent to the client
	res := Response{c.action, statusCode, statusAndData[dataIndex], c.ID}

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
	dataIndex := 0
	statusCode := 200

	// take user dataIndex from the statusAndCode and assign them to the above variables
	assignData(&dataIndex, &statusCode, statusAndData, "Send")

	//the response to be sent to the client
	res := Response{c.action, statusCode, statusAndData[dataIndex], c.ID}

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
	dataIndex := 0
	statusCode := 200

	// take user dataIndex from the statusAndCode and assign them to the above variables
	assignData(&dataIndex, &statusCode, statusAndData, "Send")

	//the response to be sent to the client
	res := Response{c.action, statusCode, statusAndData[dataIndex], c.ID}

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
