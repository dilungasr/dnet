package dnet

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dilungasr/tanzanite/types"
)

// deleteContext is for deleting the context from the hub
func deleteContext(c *Ctx) {
	if _, ok := c.hub.contexts[c]; ok {
		delete(c.hub.contexts, c)
		close(c.send)
	}

	//  check if the context is present in the rooms
	for roomName, contexts := range c.hub.rooms {
		for index, ctx := range contexts {
			if ctx == c {
				//  if ctx is not the last element in the slice
				if index != len(contexts)-1 {
					c.hub.rooms[roomName] = append(contexts[:index], contexts[index+1:]...)
					close(ctx.send)
				} else {
					// if context is the last element in the slice
					c.hub.rooms[roomName] = contexts[:index]

					// check if after removing the last element ....is there any remained elements in the rooms
					if len(c.hub.rooms[roomName]) == 0 {
						delete(c.hub.rooms, roomName)
						close(ctx.send)
					}
				}
			}
		}
	}
}

// assignData is for extracting data and statusCode from the action handler and assign them to the data and statusCode respectively
func assignData(dataIndex, statusCode *int, statusAndCode []interface{}, funcName string) {
	switch {
	case len(statusAndCode) == 2:
		// check if the first data is the code
		if !types.Is("Int", statusAndCode[0]) {
			panic("The format of the " + funcName + "() should be " + funcName + "(statusCode int, Data interface{}). You can also omit the statusCode if you want it to be OK.")
		}

		// if every thing right ....
		*dataIndex = 1
		*statusCode = statusAndCode[0].(int)
	case len(statusAndCode) == 0:
		panic("Too few " + funcName + "() arguments.")
	case len(statusAndCode) > 2:
		panic("Too many " + funcName + "() arguments.")
	}
}

// jsonSender json internally using the http.ResponseWriter
func jsonSender(w http.ResponseWriter, statusCode int, data interface{}) {
	// set the mime type
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// write the status code
	w.WriteHeader(statusCode)

	// Send the json
	json.NewEncoder(w).Encode(data)
}

// resetWriteDeadline does what it says
func resetWriteDeadline(c *Ctx) {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
}
