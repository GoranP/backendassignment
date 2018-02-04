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