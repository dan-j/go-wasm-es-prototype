//go:build js

package main

import (
	"encoding/json"
	"fmt"
	"github.com/dan-j/go-wasm/core"
	"syscall/js"
)

func main() {
	fmt.Println("hello world")
	js.Global().Set("multiply", js.FuncOf(multiply))
	js.Global().Set("handleEvent", js.FuncOf(handleEvent))
	<-make(chan struct{})
}

// This function is exported to JavaScript, so can be called using
// exports.multiply() in JavaScript.
//export multiply
func multiply(this js.Value, args []js.Value) interface{} {
	return args[0].Int() * args[1].Int()
}

// //export multiply
// func multiply(x, y int) int {
// 	return x * y
// }

//export handleEvent
func handleEvent(this js.Value, args []js.Value) interface{} {
	if len(args) != 3 {
		return "handleEvent: invalid number of args, expected 3"
	}

	state, event, cb := args[0], args[1], args[2]
	if state.Type() != js.TypeString {
		return fmt.Sprintf("state: unexpected type, want String, got %s", state.Type())
	}

	if event.Type() != js.TypeString {
		return fmt.Sprintf("state: unexpected type, want String, got %s", state.Type())
	}

	if cb.Type() != js.TypeFunction {
		return fmt.Sprintf("cb: unexpected type, want Function, got %s", cb.Type())
	}

	var thing core.Thing
	if err := json.Unmarshal([]byte(state.String()), &thing); err != nil {
		fmt.Println("state: oops: ", err)
		return nil
	}

	var e core.Event
	if err := json.Unmarshal([]byte(event.String()), &e); err != nil {
		fmt.Println("event: oops: ", err)
		return nil
	}

	if err := thing.ApplyEvent(e); err != nil {
		fmt.Println("failed to ApplyEvent(): err = ", err.Error())
		return nil
	}

	thingBytes, err := json.Marshal(thing)
	if err != nil {
		fmt.Println("failed to marshal thing JSON: ", err)
		return nil
	}

	var thingResp map[string]interface{}
	if err := json.Unmarshal(thingBytes, &thingResp); err != nil {
		fmt.Println("failed to unmarshal thingBytes into map: ", err)
		return nil
	}

	cb.Invoke(thingResp)

	return nil
}