package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"strings"
	"user-management/identity"
)

type Logic interface {
	BusinessLogic(ctx context.Context, user string, data string) (string, error)
}
type Controller struct {
	Logic Logic
}

// Login implements the worst authentication system known.
func (c Controller) Login(rw http.ResponseWriter, req *http.Request) {
	userName := req.URL.Query().Get("user")
	if len(strings.TrimSpace(userName)) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("No user specified"))
		return
	}
	identity.SetUser(userName, rw)
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("user logged in"))
}

// DoLogic is a showcase method that extracts user info from a [http.Request] context
//
//	and passes it explicitly into some business logic function
//
// NOTE:
// this code shows the value and beauty of separation of concerns;
// how the user is loaded (done by [identity.Middleware])
// is unknown to the Controller (at (1))
func (c Controller) DoLogic(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// (1) you extract the user from the request context
	user, ok := identity.UserFromContext(ctx)
	if !ok {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := req.URL.Query().Get("data")

	// (2) And pass the user explicitly into your business logic
	result, err := c.Logic.BusinessLogic(ctx, user, data)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.Write([]byte(result))
}

func (c Controller) Logout(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := identity.UserFromContext(ctx)
	if !ok {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	identity.DeleteUser(rw)
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("user logged out"))
}

type LogicImpl struct{}

func (l LogicImpl) BusinessLogic(ctx context.Context, user string, data string) (string, error) {
	return fmt.Sprintf("Hello %s, thank you for sending me %s", user, data), nil
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	controller := Controller{
		Logic: LogicImpl{},
	}
	r.Get("/login", controller.Login)
	r.Route("/business", func(r chi.Router) {
		r = r.With(identity.Middleware)
		r.Get("/", controller.DoLogic)
		r.Get("/logout", controller.Logout)
	})
	http.ListenAndServe(":3000", r)
}
