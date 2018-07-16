package types

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

func MakeVersion() string {
	return uuid.NewV4().String()
}

type VersionRecord struct {
	Data     json.RawMessage
	Version  string
	Versions []string
}

// todo talina: would it be wise to add status and plan and env
// for a layout in the same Data json?
// For the moment, I've created another struct for a layout
type LayoutRecord struct {
	Plan json.RawMessage
	Env json.RawMessage
	Status string
	Version string
	Versions []string
}

type Vars map[string]interface{}

type Namespace struct {
	Name string
	Vars *Vars
}

type Blueprint struct {
	Name string
}
