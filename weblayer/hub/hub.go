package hub

import (
	"fmt"
	"log"
	"net/http"
	"weblayer/conn"
	"weblayer/usage"
	"workerlayer/utl"

	"github.com/gorilla/websocket"
)

// serverWs handles websocket requests
func serveWs(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		utl.ERR(w, "Method not allowed", 405)
		return
	}

	//case behind proxy
	clientIP := r.Header.Get("X-Forwarded-For")

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utl.ERR("wss upgrade error", err)
		return
	}

	utl.INFO(clientIP, "serveWs", "new connection!", ws.UnderlyingConn().RemoteAddr())
	conn.StartConnection(ws)

}

func Start() {

	conn.InitRedisPool()
	port := usage.Port()
	http.HandleFunc("/ws", serveWs)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	utl.INFO("Listening plain text http/ws on port", port)

}
