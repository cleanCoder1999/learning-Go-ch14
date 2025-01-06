package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func makeRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func main() {

	// context.WithCancel() adds cancellation properties to the parent context (context.Background()) and returns a child context (ctx)
	// context.Background() ... returns an empty context
	//ctx, cancelFunc := context.WithCancel(context.Background())
	ctx, cancelFunc := context.WithCancelCause(context.Background())

	// defers the cancellation function that is associated with the created context ctx
	// NOTE:
	// Anytime you create a context that has an associated cancel function, you MUST call that cancel function
	// when you are done processing, whether or not your processing ends in an error
	//
	// If you don't, your program will leak resources (memory and goroutines).
	//
	// No error occurs if you call the cancel function more than once; any invocation after the first does nothing.
	//defer cancelFunc()
	defer cancelFunc(nil)

	ch := make(chan string)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			// ctx is passed into makeRequest()
			resp, err := makeRequest(ctx, "http://httpbin.org/status/200,200,200,500")
			if err != nil {
				// if an error occurs, the context is cancelled
				//cancelFunc()
				cancelFunc(fmt.Errorf("in status goroutine: %w", err))
				return
			}

			if resp.StatusCode == http.StatusInternalServerError {
				// if an error occurs, the context is cancelled
				//cancelFunc()
				cancelFunc(errors.New("bad status"))
				return
			}

			select {
			case ch <- "success from status":

				// to detect a cancellation, the context.Context interface has a method called Done().
				// it returns a channel of type struct{}
				// this channel is closed when the cancel function is invoked
			case <-ctx.Done():
			}

			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		defer wg.Done()

		for {
			// ctx is passed into makeRequest()
			resp, err := makeRequest(ctx, "http://httpbin.org/delay/1")
			if err != nil {
				fmt.Println("in delay goroutine:", err)
				// if an error occurs, the context is cancelled
				//cancelFunc()
				cancelFunc(fmt.Errorf("in delay goroutine: %w", err))
				return
			}

			select {
			case ch <- "success from delay: " + resp.Header.Get("date"):
			case <-ctx.Done():
			}
		}
	}()

loop:
	for {
		select {
		case s := <-ch:
			fmt.Println("in main:", s)
		case <-ctx.Done():
			fmt.Println("in main: cancelled!", context.Cause(ctx))
			break loop
		}
	}

	wg.Wait()
	fmt.Println("context cause:", context.Cause(ctx))
}
