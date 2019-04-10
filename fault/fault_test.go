package fault

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	raven "github.com/getsentry/raven-go"
	"github.com/stretchr/testify/assert"
)

var ch chan []byte

func TestMain(m *testing.M) {
	ch = make(chan []byte, 1)
	sentryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ch <- []byte("test panic bytes")
		w.WriteHeader(http.StatusOK)
	}))

	uri, _ := url.Parse(sentryServer.URL)
	uri.User = url.UserPassword("public", "secret")
	dsn := fmt.Sprintf("%s/sentry/test-project", uri)
	if err := raven.SetDSN(dsn); err != nil {
		fmt.Printf("Sentry init error %v ", err)
	}

	defer sentryServer.Close()
	m.Run()
}

func TestSentry(t *testing.T) {

	t.Run("Testing capture panic send to sentry", func(t *testing.T) {
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
		assert.Equal(t, "test panic bytes", string(<-ch))
	})
}

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
	assert.Equal(t, "test panic bytes", string(<-ch))
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
	assert.Equal(t, "test panic bytes", string(<-ch))
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
	assert.Equal(t, "test panic bytes", string(<-ch))
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
