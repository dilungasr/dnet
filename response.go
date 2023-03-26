package dnet

// response models data sent to the client.
type response struct {
	// identifies the API endpoint
	Action string `json:"action"`
	//status  code to be returned to the client
	Status int `json:"status"`
	// payload to be sent to the client
	Data interface{} `json:"data"`
	// Holds ID of the sender
	Sender string `json:"sender"`

	IsSource bool `json:"isSource"`
}

func (res response) checkSource(source1, source2 unionContext) response {
	res.IsSource = source1 == source2

	return res
}

func newResponse(c unionContext, status int, data interface{}) response {
	return response{
		Action:   c.getAction(),
		Status:   status,
		Data:     data,
		Sender:   c.getID(),
		IsSource: false,
	}
}
