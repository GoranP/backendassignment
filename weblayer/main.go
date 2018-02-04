package main

import "weblayer/hub"

/*
COMMANDS

{"Cmd":1,"CmdData":{"UserName":"branko","FavoriteNumber":11}}
{"Cmd":1,"CmdData":{"UserName":"marko","FavoriteNumber":7}}
{"Cmd":1,"CmdData":{"UserName":"ana","FavoriteNumber":22}}
{"Cmd":1,"CmdData":{"UserName":"mihaela","FavoriteNumber":66}}

{"Cmd":2,"CmdData":{"UserName":"ana"}}

*/

func main() {
	hub.Start()
}
