package conn

import (
	"fmt"
	"time"
	"weblayer/usage"
	"workerlayer/utl"

	"github.com/garyburd/redigo/redis"
)

///////redis stuff

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

func InitRedisPool() {
	//connect to redis && create pool
	redisURL := fmt.Sprintf("redis://%s:6379", usage.Redis())
	utl.INFO("Connecting to redis ->", redisURL)
	Pool = newPool(redisURL)
}

// connection will communicate with backend worker over channel conn.{connid} and worker.{connid}
// connid in this case is uuid created in StartConnection()
func (c *Connection) sendToRedis(m []byte) {
	//send to channel "conn.{connid}""
	rc := Pool.Get()
	defer rc.Close()

	err := rc.Send("PUBLISH", fmt.Sprintf("conn.%s", c.id), string(m))
	if err != nil {
		utl.ERR("sendToRedis PUBLISH", err)
		return
	}
	err = rc.Flush()
	if err != nil {
		utl.ERR("sendToRedis flush", err)
		return
	}

}

// separate go routine just to read events from redis sent from backend woekre process
// go routine will exit when connection to redis is closed
// connection to redis is closed when websocket is closed
func (c *Connection) readWorkerMessages() {
	//subscribe on channel "worker.{connid}""
	rc := Pool.Get()
	//defer rc.Close()
	defer utl.INFO("defer stop readWorkerMessage")

	c.psc = redis.PubSubConn{Conn: rc}

	utl.INFO(fmt.Sprintf("subscribe on redis channel -> worker.%s", c.id))
	err := c.psc.Subscribe(fmt.Sprintf("worker.%s", c.id))
	if err != nil {
		utl.ERR("SUBSCRIBE", err)
		return
	}
	rc.Flush()
	for {
		// process messages from redis
		reply := c.psc.Receive()
		switch n := reply.(type) {
		case error:
			utl.INFO("stop readWorkerMessage")
			return
		case redis.Message:
			utl.INFO("reply from worker:", n.Channel, string(n.Data))
			//send date to client
			c.Send <- n.Data
		case redis.Subscription:
			utl.INFO("Un/Subscription message -->", n.Channel, n.Kind)
			if n.Kind == "unsubscribe" {
				return
			}
		}
	}
}
