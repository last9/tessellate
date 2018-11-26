package fault

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, expected, actual, fmt.Sprintf("Expected %v Got %v", expected, actual))
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
	assert.Equal(t, expected, actual, fmt.Sprintf("Expected %v Got %v", expected, actual))
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
	assert.Equal(t, expected, actual, fmt.Sprintf("Expected %v Got %v", expected, actual))
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
		assert.NotEqual(t, 1, counter, fmt.Sprintf("Fault handler should not have been called, counter %+v", counter))
	}
}
