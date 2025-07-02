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

	AsyncID string `json:"asyncId"`
}

func (res response) checkSource(source1, source2 unionContext) response {
	res.IsSource = source1 == source2

	if res.IsSource {
		res.AsyncID = source1.getAsyncID()
	}

	return res
}

func (res *response) setSource(ctx *Ctx) {
	res.IsSource = true
	res.AsyncID = ctx.asyncID
}

func newResponse(c unionContext, status int, data interface{}, isOriginalAction ...bool) response {
	if len(isOriginalAction) == 0 {
		isOriginalAction = []bool{false}
	}

	action := c.GetAction()

	if isOriginalAction[0] {
		action = c.GetOriginalAction()
	}

	return response{
		Action:   action,
		Status:   status,
		Data:     data,
		Sender:   c.getID(),
		IsSource: false,
	}
}
