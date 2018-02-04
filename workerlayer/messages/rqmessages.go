package messages

//////////////////////////////////////////////////////////////////////////
// messages sent from client  - requests (either set, update or get data)
//////////////////////////////////////////////////////////////////////////

//go:generate stringer -type=RQEnum
type RQEnum int

const (
	RQUnknown RQEnum = iota
	RQSetFavoriteNumber
	RQListAllUsers
)

//interface for client messages
type ClientRequest interface {
	Request() RQEnum
}

//small struct to indentify message, base "class" for other messages
//implements ClientRequest interface
type ClientRQ struct {
	Cmd RQEnum
}

func (c ClientRQ) Request() RQEnum {
	return c.Cmd
}

/* COMMANDS
{"Cmd":1,"CmdData":{"UserName":"branko","FavoriteNumber":11}}
{"Cmd":1,"CmdData":{"UserName":"marko","FavoriteNumber":7}}
{"Cmd":1,"CmdData":{"UserName":"ana","FavoriteNumber":22}}
{"Cmd":1,"CmdData":{"UserName":"mihaela","FavoriteNumber":66}}

{"Cmd":2,"CmdData":{"UserName":"ana"}}
*/

//1. a message to set a user's favorite number
type ClientSetFavoriteNumber struct {
	ClientRQ
	CmdData SetFavoriteNumber
}

type SetFavoriteNumber struct {
	UserName       string
	FavoriteNumber int
}

//2. a message to list all users (sorted alphabetically) and their favorite numbers
type ClientGetList struct {
	ClientRQ
	CmdData GetList
}

type GetList struct {
	UserName string //username as a identifier
}
