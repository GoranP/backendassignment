package messages

//go:generate stringer -type=RPEnum
type RPEnum int

// server messages/replies
const (
	Empty RPEnum = iota
	SrvListAllUsers
)

type User struct {
	Username string
	Favnum   int
}

//general reply structure for API commmand
type Reply struct {
	Cmd    RPEnum
	Status string //OK or NOTOK
	Error  string
}

type AllUserlist struct {
	Reply
	AllUsers []User
}
