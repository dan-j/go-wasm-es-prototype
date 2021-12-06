# Go WASM Event Sourcing Prototype

Prototype application demonstrating using common GoLang code on the server and in browser via WebAssembly.

The general premise is our data is event sourced, the backend holds an event store of all messages sent from the client.
On an initial websocket connection, the server computes the current state by applying all events on top of one another 
and returns this to the client. As new events are received from a client, they are streamed back to all other clients. 
The client then uses the same backend code for computing current state via the WASM module.

A success demo of this application involves:

1. Two browser tabs connected to the WS server; both tabs show current state
2. One tab submits some events
3. Both tabs should be updated with the current state

Additionally, the clients should only ever receive the state object once, on initial connection

## How to run

```sh
GOOS=js GOARCH=wasm go build -o html/wasm.wasm ./wasm/main.go
go run main.go
```

Open two browser tabs at `http://localhost:8080`

In one tab, interact with the forms and buttons, switch to the other tab and watch it update!