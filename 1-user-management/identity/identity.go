package identity

import (
	"context"
	"net/http"
)

// Two patterns are used to guarantee that a key is unique and comparable
// (1) unexported type based on int with an unexported constant of that type
// (2) unexported type using an empty struct
//
// How do you know which key-style to use?
//
// If you have ...
// (1) a set of related keys for storing different values in the context, use the int and iota technique
// (2) only a single key, either is fine
//
// The important thing is that you want to make it impossible for context keys to collide (with ones from other packages)

// userKeyStruct is one key-style option to define an unexported key type
type userKeyStruct struct{}

// NOTE: the name of the function that creates a context should start with "ContextWith"
func ContextWithUserStruct(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, userKeyStruct{}, user)
}

// NOTE: the name of the function that returns the value from the context should have a name that ends with "FromContext"
func UserFromContextStruct(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(userKeyStruct{}).(string)
	return user, ok
}

// userKey is the second key-style option to define an unexported key type
type userKey int

const (
	// in Go, it is a common pattern to assign the first iota value in the constant block to
	// _
	// or
	// to a constant value that indicates the value is invalid
	//
	// this makes it easy to detect that a variable has not been properly initialized
	_ userKey = iota
	key
)

// NOTE: the name of the function that creates a context should start with "ContextWith"
func ContextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, key, user)
}

// NOTE: the name of the function that returns the value from the context should have a name that ends with "FromContext"
func UserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(key).(string)
	return user, ok
}

// a real implementation would be signed to make sure
// the identity didn't spoof their identity
func extractUser(req *http.Request) (string, error) {
	userCookie, err := req.Cookie("identity")
	if err != nil {
		return "", err
	}
	return userCookie.Value, nil
}

// Middleware defines how user information is loaded / managed
func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		user, err := extractUser(req)
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("unauthorized"))
			return
		}

		ctx := req.Context()
		// creates a new context that contains the user
		ctx = ContextWithUser(ctx, user)
		// creates a new request with the old request and the context containing the user
		req = req.WithContext(ctx)
		h.ServeHTTP(rw, req)
	})
}

func SetUser(user string, rw http.ResponseWriter) {
	http.SetCookie(rw, &http.Cookie{
		Name:  "identity",
		Value: user,
	})
}

func DeleteUser(rw http.ResponseWriter) {
	http.SetCookie(rw, &http.Cookie{
		Name:   "identity",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}
