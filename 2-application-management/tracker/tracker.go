// Package tracker showcases some application management implementation that allows to track request information
//
//	along a request chain, and across different services
//
// By using the dependency injection technique with implicit interfaces,
// any business logic is completely unaware of any tracking information.
// see [main.Logger] and [main.RequestDecorator]
package tracker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// guidKey represents an unexported key-type for writing and reading a [uuid] into a [context.Context]
type guidKey int

// key is an unexported constant of the key-type guidKey
const key guidKey = 1

// contextWithGUID is an API to add a [uuid] to a [context.Context]
//
// it takes an existing [context.Context] and a [uuid]
// and combines those to create a new [context.Context]
func contextWithGUID(ctx context.Context, guid string) context.Context {
	return context.WithValue(ctx, key, guid)
}

// contextWithGUID is an API to read a [uuid] from a [context.Context] instance
func guidFromContext(ctx context.Context) (string, bool) {
	g, ok := ctx.Value(key).(string)
	return g, ok
}

func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		if guid := req.Header.Get("X-GUID"); guid != "" {
			// if a [uuid] is in the request, it is written into the context
			ctx = contextWithGUID(ctx, guid)
		} else {
			// if no [uuid] is in the request, a new one is created and written into the context
			ctx = contextWithGUID(ctx, uuid.New().String())
		}

		// the old request and the enriched context are used to create a new request
		req = req.WithContext(ctx)
		h.ServeHTTP(rw, req)
	})
}

type Logger struct{}

// Log offers a generic logging method that takes in a [context.Context] and a string
//
// if there is a [uuid] in the context it appends it to the beginning of the log message
// and outputs it
func (Logger) Log(ctx context.Context, message string) {
	if guid, ok := guidFromContext(ctx); ok {
		message = fmt.Sprintf("GUID: %s - %s", guid, message)
	}
	// do logging
	fmt.Println(message)
}

// Request is used when this service makes a call to another service
//
// it takes in an [*http.Request], adds the header with the [uuid]
// if it exists in the [context.Context] instance, and
// returns the [*http.Request]
func Request(req *http.Request) *http.Request {
	ctx := req.Context()
	if guid, ok := guidFromContext(ctx); ok {
		req.Header.Add("X-GUID", guid)
	}
	return req
}
