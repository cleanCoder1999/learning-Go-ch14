package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"
)

type keyType int

const (
	// in Go, it is a common pattern to assign the first iota value in the constant block to
	// _
	// or
	// to a constant value that indicates the value is invalid
	//
	// this makes it easy to detect that a variable has not been properly initialized
	_ keyType = iota
	key
)

type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
)

func main() {
	// ### - exercise 2: write a program that adds randomly generated numbers between 0 and 100_000_000 together until 1234 is generated or 2 seconds have passed
	{
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(2)*time.Second)
		defer cancelFunc()

		reason := "number reached"
		var i, sum, v int
		for ; v != 1234; v = rand.IntN(100_000_000) {
			sum += v

			if err := context.Cause(ctx); err != nil {
				reason = "timeout"
				break
			}
			i++
		}

		fmt.Println("termination cause:", reason)
		fmt.Println("number of iterations:", i)
		fmt.Printf("sum: %d\n\n", sum)
	}

	// ### - exercise 3: advanced logging with contexts
	{
		mux := http.NewServeMux()
		mux.Handle("/log", ExtractLogLevelMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Log(r.Context(), Debug, "the log level is set and everything works properly")
		})))

		s := http.Server{
			Addr:         ":8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 90 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      mux,
		}

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := s.ListenAndServe()
			if err != nil {
				log.Fatal(err)
			}
		}()

		wg.Wait()
	}
}

// ### - exercise 3
func Log(ctx context.Context, level Level, msg string) {

	inLevel, ok := LogLevelFromContext(ctx)
	if !ok {
		slog.Warn("no log level available")
	}

	if level == Debug && inLevel == Debug {
		slog.Info(msg)
	}
	if level == Info && (inLevel == Debug || inLevel == Info) {
		slog.Info(msg)
	}
}

// LogLevelFromContext extracts the log level from the context, if it is available
func LogLevelFromContext(ctx context.Context) (Level, bool) {
	v, ok := ctx.Value(key).(Level)
	return v, ok
}

// ContextWithLogLevel stores the log level in a context that is derived from its parent context passed into the function
func ContextWithLogLevel(ctx context.Context, l Level) context.Context {
	return context.WithValue(ctx, key, l)
}

func ExtractLogLevelMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var logLevel Level

		param := r.URL.Query().Get("log_level")

		switch param {
		case string(Debug):
			logLevel = Debug
		case string(Info):
			logLevel = Info
		default:
			slog.Warn("invalid param value:", param)
		}

		// if logLevel is set, create a new request that contains a context with the logLevel assigned
		if logLevel != "" {
			ctx := r.Context()
			ctx = ContextWithLogLevel(ctx, logLevel)
			r = r.WithContext(ctx)
		}

		h.ServeHTTP(w, r)
	})
}

// ### - exercise 1: middleware-generating function that creates a context with a timeout; return func(http.Handler) http.Handler
func Timeout(ms int) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx := r.Context()
			ctx, cancelFunc := context.WithTimeout(ctx, time.Duration(ms)*time.Millisecond)
			defer cancelFunc()

			h.ServeHTTP(w, r)
		})
	}
}
