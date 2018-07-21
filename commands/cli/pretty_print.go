package main

import (
	"encoding/json"
	"log"
)

func prettyPrint(w interface{}) {
	b, _ := json.MarshalIndent(w, "", "    ")
	log.Println(string(b))
}
