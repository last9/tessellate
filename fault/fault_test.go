package fault

import (
	"errors"
	"sync"
	"testing"
)

func TestFaultHandlerIntError(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	expected := "Crash: 1"
	var actual string

	go func() {
		defer Handler(func(err error) {
			actual = err.Error()
			wg.Done()
		})

		panic(1)
	}()

	wg.Wait()
	if expected != actual {
		t.Errorf("Expected %v Got %v", expected, actual)
	}
}

func TestFaultHandlerStringError(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	expected := "Crash: new"
	var actual string

	go func() {
		defer Handler(func(err error) {
			actual = err.Error()
			wg.Done()
		})

		panic(errors.New("new"))
	}()

	wg.Wait()
	if expected != actual {
		t.Errorf("Expected %v Got %v", expected, actual)
	}
}

func TestFaultHandlerError(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	expected := "Crash: hello world"
	var actual string

	go func() {
		defer Handler(func(err error) {
			actual = err.Error()
			wg.Done()
		})

		panic("hello world")
	}()

	wg.Wait()
	if expected != actual {
		t.Errorf("Expected %v Got %v", expected, actual)
	}
}

func TestFaultHandlerNoError(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	var counter int

	go func() {
		defer wg.Done()
		defer Handler(func(err error) {
			counter++
		})
		counter++
		return
	}()

	wg.Wait()
	if counter != 1 {
		t.Log("Counter", counter)
		t.Errorf("Fault handler should not have been called")
	}
}
