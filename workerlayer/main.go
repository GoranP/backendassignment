package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"workerlayer/messages"
	"workerlayer/unmarshall"
	"workerlayer/usage"
	"workerlayer/utl"

	"github.com/garyburd/redigo/redis"
)

var pscmap map[string]redis.PubSubConn
var mux sync.Mutex

func main() {
	pscmap = make(map[string]redis.PubSubConn, 0)
	initRedis()
	readConnMessages()
}

//subscribe on channel "conn.*""
func readConnMessages() {

	rc := Pool.Get()
	defer rc.Close()
	psc := redis.PubSubConn{Conn: rc}

	err := psc.PSubscribe("conn.*")
	if err != nil {
		utl.ERR("psub", err)
	}
	utl.INFO("worker subscribed on redis channel conn.*")
	for {
		reply := psc.Receive()
		if err != nil {
			utl.ERR("Recieve failed", err)
			return
		}
		//utl.WARN(reply)
		// process pushed message
		switch n := reply.(type) {
		case error:
			utl.ERR("reply err", err)
			return
		case redis.Message:
			utl.INFO("Standard message:", n.Channel, string(n.Data))
		case redis.Subscription:
			utl.INFO("Un/Subscription message-->", n.Kind, n.Channel)
		case redis.PMessage:
			//scale - run processMessage in separate go routine
			go processMessage(n)
		}
	}
}

func processMessage(n redis.PMessage) {

	utl.INFO("Process message", "channel:", n.Channel, "data:", string(n.Data))

	s := strings.Split(n.Channel, ".")
	id := s[1]
	//special case message - init
	//little hack to init channel and to send message back to connection that has not sent anything to the worker, only latently connected
	if string(n.Data) == "init" {
		go pushKeyChanges(id)
		return
	}
	//special case message - closed
	//whn client lost connection - kill go routines
	if string(n.Data) == "closed" {

		mux.Lock()
		if sc, exists := pscmap[id]; exists {
			sc.PUnsubscribe("*keyspace*:user:*")
			delete(pscmap, id)
		}
		mux.Unlock()
		return
	}

	request := unmarshall.Unmarshall(n.Data)
	if request == nil {
		utl.ERR("invalid json")
		return
	}

	switch request.Request() {
	case messages.RQSetFavoriteNumber:
		//update number
		cmd := request.(messages.ClientSetFavoriteNumber)
		cmdData := cmd.CmdData
		setData(cmdData)
	case messages.RQListAllUsers:
		//notify all users connections with sorted list
		//extract ID from channel name
		rpl := Pool.Get()
		rpl.Send("PUBLISH", "worker."+id, string(utl.JSON(getAllUsers())))
		rpl.Flush()
		rpl.Close()
	}

}

//subscribe on redis events when change happens on keys
//in case of change send all users and favorite numbers to client, over channel
func pushKeyChanges(id string) {

	rc := Pool.Get()
	//defer rc.Close()

	defer utl.INFO("pushKey goroutine ended", id)

	psc := redis.PubSubConn{Conn: rc}

	//store conn in map
	mux.Lock()
	pscmap[id] = psc
	mux.Unlock()

	err := psc.PSubscribe("*keyspace*:user:*")
	if err != nil {
		utl.ERR("pushKeyChanges", err)
		return
	}
	utl.INFO("Subscribe on *keyspace*:user:*")
	for {
		reply := psc.Receive()
		if err != nil {
			utl.ERR("Recieve failed", err)
			return
		}
		// process pushed message
		switch n := reply.(type) {
		case error:
			return
		case redis.Message:
		case redis.Subscription:
			if n.Kind == "punsubscribe" {
				return
			}
		case redis.PMessage:
			rpl := Pool.Get()
			rpl.Send("PUBLISH", "worker."+id, string(utl.JSON(getAllUsers())))
			rpl.Flush()
			rpl.Close()
		}
	}
}

func setData(data messages.SetFavoriteNumber) {
	rc := Pool.Get()
	defer rc.Close()

	namekey := fmt.Sprintf("user:%s", data.UserName)
	rc.Send("HMSET", namekey, "username", data.UserName, "favnum", data.FavoriteNumber)
	rc.Send("SADD", "users", data.UserName)
	rc.Flush()
}

func getAllUsers() []messages.User {
	rc := Pool.Get()
	defer rc.Close()

	var users []messages.User

	utl.INFO("get values")
	values, err := redis.Values(rc.Do("SORT", "users", "ALPHA",
		"BY", "user:*->username",
		"GET", "user:*->username",
		"GET", "user:*->favnum"))
	if err != nil {
		utl.ERR(err)
		return users
	}

	if err := redis.ScanSlice(values, &users); err != nil {
		utl.ERR("ScanSlice - getallusers error", err)
		return users
	}
	return users
}

/////////////////redis conn pool
///////////////////////////////////
var Pool *redis.Pool

func newPool(redisURL string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(redisURL) },
	}
}

func initRedis() {
	//connect to redis && create pool
	redisURL := fmt.Sprintf("redis://%s:6379", usage.Redis())
	Pool = newPool(redisURL)

}
