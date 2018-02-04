package unmarshall

import (
	"encoding/json"
	"workerlayer/messages"
	"workerlayer/utl"
)

//////////////////////////////
//Helper unmarshall functions

//unmarshall message from client with smarat unmarshall function
func Unmarshall(message []byte) (gameCmd messages.ClientRequest) {

	command := smartUnmarshall(message)
	if command == nil {
		utl.ERR("connection.Unmarshal()", "could not unmarshall string in command --> ", string(message))
	}
	return command
}

//unmarshall message based on enum in message
//"smart" unmarshall.... not so smart at all
func smartUnmarshall(message []byte) messages.ClientRequest {

	ccmd := messages.ClientRQ{}
	err := json.Unmarshal(message, &ccmd)
	if err != nil {
		utl.INFO("smartUnmarshall()", "json error:", err.Error())
		return nil
	}
	switch ccmd.Cmd {

	case messages.RQSetFavoriteNumber:
		cmd := messages.ClientSetFavoriteNumber{}
		err = json.Unmarshal(message, &cmd)
		if err != nil {
			return nil
		}
		return cmd
	case messages.RQListAllUsers:
		cmd := messages.ClientGetList{}
		err = json.Unmarshal(message, &cmd)
		if err != nil {
			return nil
		}
		return cmd
	case messages.RQUnknown:
		utl.ERR("Unknown command - not initialized structs on client")
	default:
		utl.ERR("smartUnmarshall()", "Unknown command", ccmd.Request())

	}
	return nil

}
