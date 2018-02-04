# backendassignment

## Backend Engineer Assignment

Every user (identified by a username) has a favorite number. The system maintains a list of users and their favorite numbers, and exposes a single websocket endpoint. The websocket accepts two types of messages:

1. a message to set a user's favorite number
2. a message to list all users (sorted alphabetically) and their favorite numbers

The websocket has one type of response message:

1. the alphabetical listing of all known users and their favorite numbers

While a websocket connection is active, the response message should be delivered to it when any user's favorite number is changed, even by another connection.

The system has three layers:

1. the web layer, accepting and handling websocket connections
2. the broker and data store layer, implemented by Redis
3. the worker layer, handling the request messages and generating the responses

When layer 1 receives a request, it should forward it to layer 3 over layer 2. Layer 3 should respond over layer 2. Layers 1 and 3 should never communicate directly. Each layer should be viewed as a separate component.

Define simple messages that fulfill the requirements. Implement layers 1 and 3 using `Python` or `Go`. Bonus points if implemented in both languages (and they can interoperate). Use the official `Redis` `Docker` image for layer 2. If you think it's needed, include a simple HTML page with layer 1 that can be used to send requests and display responses in the browser. Otherwise provide `curl/wsta` commands to send requests.

All components should be `Dockerized` (come with a `Dockerfile`) and a `Docker Compose` configuration file should be included to start and link the entire system.

Also consider the infrastructure part and deploy your code somewhere in the cloud (`AWS`, `Azure`, `Google App Engine` ...). Think about how you could easily (someday) scale your implementation. 

Provide your implementation in a public `Github` repo and send the `URL` once you're finished with it.

You're free to add any additional whiz, magic or what you see fit.

Direct any questions to `serverdev@nanobit.co`

## Solution

Golang should be installed and GOPATH should point to the root of this project.
Scripts and code are developed for Linux.

Redis is configured to fire events upon key modification (notify-keyspace-events KEA).

Weblayer and workerlayer communicate through Redis Pub/Sub channels.
Weblayer just publishes incoming JSON's to `conn.{connid}` channel. Each websocket gets it own goroutines that handles communication.
Weblayer subscribes to `worker.{connid}` channel.


Workerlayer (p)subscribes to `conn.*` and receive all JSON's from clients. JSON's are unmarshaled, and depending on message either user's fav number is updated/created or a sorted list of all users is retrieved from Redis and sent to the client over `worker.{connid}` channel. Sorting is done on Redis via two keys: HASH (user:xx) and Set (users). 
`SORT users ALPHA BY user:*->username GET user:*->username GET user:*->favnum`


Worklayer starts separate goroutine per client 'pushKeyChanges' that subscribes to event `*keyspace*:user:*`. When an event is fired, then all clients get a latest sorted list of users and favnumbers.

Both components are completely independent and horizontally scalable. Hardpoint for both components is Redis server.

Everything is dockerized and to start the system execute docker/dockerbuild shell script.

In dockerbuild script is commented code as an example of docker push to AWS. Read more info how to test deploy in the script itself.

Examples of JSON's

First start wsta:
`wsta ws://localhost:9999/ws`

-add some data to redis (copy/paste)


`{"Cmd":1,"CmdData":{"UserName":"branko","FavoriteNumber":11}}`

`{"Cmd":1,"CmdData":{"UserName":"marko","FavoriteNumber":7}}`

`{"Cmd":1,"CmdData":{"UserName":"ana","FavoriteNumber":22}}`

`{"Cmd":1,"CmdData":{"UserName":"mihaela","FavoriteNumber":66}}`

-get sorted list of users and fav numbers (UserName is not important and can be ommited)

`{"Cmd":2,"CmdData":{"UserName":"ana"}}`
