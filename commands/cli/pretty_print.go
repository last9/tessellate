package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func prettyPrint(w interface{}) {
	b, _ := json.MarshalIndent(w, "", "    ")
	fmt.Fprintf(os.Stdout, string(b)+"\n")
}
