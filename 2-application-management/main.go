package main

import (
	"application-management/tracker"
	"context"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

type Logic interface {
	Process(ctx context.Context, data string) (string, error)
}
type Controller struct {
	Logic Logic
}

func (c Controller) First(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	data := req.URL.Query().Get("data")
	result, err := c.Logic.Process(ctx, data)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.Write([]byte(result))
}

// Logger offers an interface that implicitly matches [tracker.Logger.Log]
type Logger interface {
	Log(context.Context, string)
}

// RequestDecorator implicitly matches [tracker.Request] and allows to wire up a dependency that only main() is aware of
type RequestDecorator func(*http.Request) *http.Request

type LogicImpl struct {
	RequestDecorator RequestDecorator
	Logger           Logger
	Remote           string
}

func (l LogicImpl) Process(ctx context.Context, data string) (string, error) {
	l.Logger.Log(ctx, "starting Process with "+data)
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet, l.Remote+"/second?query="+data, nil)
	if err != nil {
		l.Logger.Log(ctx, "error building remote request:"+err.Error())
		return "", err
	}
	req = l.RequestDecorator(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Logger.Log(ctx, "error building remote request:"+err.Error())
		return "", err
	}
	if resp.Body == nil {
		l.Logger.Log(ctx, "empty response from second")
		return "", nil
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	return string(out), err
}

func main() {
	r := chi.NewRouter()
	r.Use(tracker.Middleware)
	controller := Controller{

		// NOTE:
		// the GUID is passed through to the logger and request decorator without the business logic being aware of it,
		//
		// SEPARATING
		// (1) the DATA needed FOR PROGRAM LOGIC
		// FROM
		// (2) the DATA needed FOR PROGRAM MANAGEMENT
		//
		// the only place that's aware of the association is the code below that wires up the dependencies
		Logic: LogicImpl{
			RequestDecorator: tracker.Request,
			Logger:           tracker.Logger{},
			Remote:           "http://localhost:4000",
		},
	}
	r.Get("/first", controller.First)
	http.ListenAndServe(":3000", r)
}
