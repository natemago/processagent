package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Message holds the message data received as JSON.
type Message struct {
	Name string `json:"name"`
}

func main() {
	// Read all of the STDIN input.
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	// Then unmarshal the JSON into our structure.
	message := &Message{}
	if err = json.Unmarshal(data, message); err != nil {
		panic(err)
	}

	// Print out the result on stdout.
	fmt.Printf("Hello %s! This is service.", message.Name)
}
