package main

import (
	"encoding/json"
	"fmt"
)

func prettyPrint(w interface{}) {
	b, _ := json.MarshalIndent(w, "", "    ")
	fmt.Println(string(b))
}
