package fault

import (
	"log"

	"github.com/pkg/errors"
)

func Printer(tIn interface{}) error {
	if tIn == nil {
		return nil
	}

	var rerr error

	switch t := tIn.(type) {
	case string:
		rerr = errors.New(t)
	case interface{}:
		if _err, ok := tIn.(error); ok == true {
			rerr = _err
		}
	}

	if rerr == nil {
		rerr = errors.Errorf("%+v", tIn)
	}

	log.Printf("%+v\n", rerr)
	return rerr
}

// Handler which accepts an on-erro routine to be called,
// after its done catching the panic.
func Handler(onErr func(err error)) {
	err := recover()
	if err == nil {
		return
	}
	onErr(errors.Wrap(Printer(err), "Crash"))
}
