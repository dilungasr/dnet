package dnet

// Map models json data to be sent to the client by creating a map[string]interface{}
type Map map[string]interface{}

// Msg models the msg to a simple msg:string map
func Msg(msg string) (simpleMsgResponse Map) {
	return Map{"msg": msg}
}
