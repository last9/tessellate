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

const (
	INACTIVE = iota  // a == 1 (iota has been reset)
	ACTIVE = iota  // b == 2
)

type LayoutRecord struct {
	Id string
	Plan json.RawMessage
	Vars *Vars
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
