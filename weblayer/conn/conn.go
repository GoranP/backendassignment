// conn package handles averything related to permanent websocket with client
// this includes:

// * reading messages and send them over radis to worker
// * writing messages to client over permanent websocket
// * managing size of max message
// * ping and managing socket alive with client
package conn

import (
	"fmt"
	"time"
	"workerlayer/utl"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
)

//////////////////////////////////////////////////////
// Connection struct encapsulate websocket connection
//
// Controller go routine controls message flow and lifecycle of read and write goroutines of websocket
// this includes pinging client, act as proxy toward read and write go routines
// most important task is to handle clean exit of all goroutines
// that way there is no "goroutine leak" that could be issue with large number of concurrent connections

// architecture:

// 	-> go read() 					-> go write()
//  (read from web socket)		    (write to websocket)
//               |                       |
//               |                       |
//               | (reader chan)         | (writer chan)
//               |                       |
//               |->     controller() --
// 		(get messages from other routines, send message to write routine, read messages from read routine)
//                       / \
//                      | c |
//                      | h |
//                      | a |
//                      | n |
//                       \ /
// 				 *other go routines*
//				- through c.Send channel send msg to client
//				- through hub.hubchannel channel get msg from client
//

const (
	// time period for auth check
	authWait = 10 * time.Second
	// delay ending of connection - for eventual messages from other routines before exiting
	connDelay = 2 * time.Second
	// Time allowed to write a message to the peer.
	writeWait = 50 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 600 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 7) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 15
)

//wrapper over websocket
type Connection struct {
	// The websocket connection.
	ws *websocket.Conn
	//uuid
	id string

	//connection subscribed on redis channel for this websocket connection
	psc redis.PubSubConn

	//controller helper channels
	reader chan []byte
	writer chan []byte

	//channels for read or write go routine to notify controller
	readerror  chan error
	writeerror chan error
	//controller sends ping rythm to write goroutine
	pinger chan bool
	// Buffered channel of outbound messages.
	Send chan []byte
	//explicitly close connection
	Close chan bool
}

/////////////////////
//CONNECTION METHODS

//read pumps messages from the websocket connection to the controller.
func (c *Connection) read() {

	msg := fmt.Sprintf("conn.read() %v %v", c.RemoteAddr(), "anonymous")
	defer utl.HandleDefer(msg, time.Now(), nil)

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			utl.WARN(c.RemoteAddr(), "read", err.Error())
			//notify controller
			c.readerror <- err
			return //this exits go routine
		}
		//send to controller, controller will send to hub
		c.reader <- message
	}
}

// write writes a message with the given message type and payload.
func (c *Connection) writeSocket(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// write pumps messages from the controller to the websocket connection.
func (c *Connection) write() {
	msg := fmt.Sprintf("conn.write() %v %v", c.RemoteAddr(), "anonymous")
	defer utl.HandleDefer(msg, time.Now(), nil)

	for {
		select {
		//controller sends message
		case message, ok := <-c.writer:
			if !ok {
				utl.INFO(c.ws.RemoteAddr().String(), "write", "send message channell - NOTOK!")
				return //this will call defer block
			}

			if err := c.writeSocket(websocket.BinaryMessage, message); err != nil {
				utl.WARN(c.ws.RemoteAddr().String(), "write", "write socket not OK.", err.Error())
				c.writeerror <- err
				return
			}

		//controller sends ping to client
		case <-c.pinger:
			//utl.Log(c.ws.RemoteAddr().String(),  "write", "Sending message PING")
			if err := c.writeSocket(websocket.PingMessage, []byte{}); err != nil {
				utl.WARN(c.ws.RemoteAddr().String(), "write", "pinger sending ping failed. ", err.Error())
				c.writeerror <- err
				return
			}
		}

	}
}

//controls message flow and lifecycle of read and write goroutines
//this includes pinging client, act as proxy toward hub,
//and proxy toward read and write go routines
//handles clean exit of all goroutines
func (c *Connection) controller() {
	//tickers

	pinger := time.NewTicker(pingPeriod)
	delay := &time.Ticker{}

	defer func() {
		if r := recover(); r != nil {
			utl.LogRecover("controller", r)
		}
		//close websocket
		c.ws.Close()
		//stop tickers
		pinger.Stop()
		delay.Stop()
		// notify worker that client websocket is closed
		c.sendToRedis([]byte("closed"))
		utl.INFO("After send to tredis")
		//unsubscribe conn from redis
		c.psc.Unsubscribe(fmt.Sprintf("worker.%s", c.id))

		utl.INFO(c.RemoteAddr(), "exiting connection controller - end of connection go routines")
	}()

	closed := false
	for {
		//when both routine signaled end - close connection and exit
		//switch for communitating with other go routines
		select {

		case message := <-c.Send:
			//send message to client through write go routine
			if !closed {
				c.writer <- message
			}
		case <-c.Close:
			if !closed {
				utl.INFO(c.ws.RemoteAddr().String(), "connController", "forced closing connection.")
				//closing socket will cause exiting of read/write pump go routines through helper channels
				CloseWS(c)
				closed = true
				delay = time.NewTicker(connDelay)
			}

		case msg := <-c.reader:
			//reader got message from client - send to worker over redis
			if !closed {
				c.sendToRedis(msg)
			}
		case <-pinger.C:
			if !closed {
				c.pinger <- true
			}
		case <-c.writeerror:
			if !closed {
				//stop reader pump when writer is stopped
				CloseWS(c)
				closed = true
				//delay of exit for eventual messages from other goroutines that use c.Send and c.Auth channels
				delay = time.NewTicker(connDelay)
			}
		case <-c.readerror:
			if !closed {
				//stop writer pump whe reader is stopped
				closed = true
				close(c.writer)
				//delay of exit for eventual messages from other goroutines that use c.Send and c.Auth channels
				delay = time.NewTicker(connDelay)
			}
		case <-delay.C:
			//exiting connection controller
			return
		}
	}

}

///////////////////
//HELPER FUNCTIONS

//send close signal to client
func CloseWS(c *Connection) {
	utl.INFO(c.ws.RemoteAddr().String(), "closeWS", "closing websocket connection.")
	c.ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server forced closed connection."), time.Now().Add(time.Second))
}

//helper for address of conection
func (c *Connection) RemoteAddr() string {
	//return c.ws.RemoteAddr().String()
	return fmt.Sprintf("%s", c.ws.RemoteAddr().String())
}

//set websocket options and PONG handler
func (c *Connection) wsOptions() {
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(
		func(string) error {
			c.ws.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
}

/////////////////////////////////////////////
////// starter function - connection factory
/////////////////////////////////////////////
//create new connection, initialize channles, starts goroutines
func StartConnection(ws *websocket.Conn) {

	c := &Connection{ws: ws}

	c.id = uuid.New()
	c.Send = make(chan []byte)

	c.reader = make(chan []byte)
	c.writer = make(chan []byte)

	c.pinger = make(chan bool)

	c.readerror = make(chan error)
	c.writeerror = make(chan error)

	//setup read options
	c.wsOptions()
	//start two go routines to read and write separatley
	go c.write()
	go c.read()
	//controller controlls message flow and lifecycle through read and write goroutines
	go c.controller()

	//start redis routine to catch message from worker
	go c.readWorkerMessages()
	//special command
	c.sendToRedis([]byte("init"))
	//new connection - send stat
	utl.INFO("connection++")
}
