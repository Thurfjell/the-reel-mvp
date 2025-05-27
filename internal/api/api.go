package api

import (
	"context"
	"log"
	"net/http"
	"time"
)

type RouteMeta struct {
	Path    string
	Handler http.Handler
}

type Api struct {
	Server *http.Server
}

func (api *Api) Run() {
	go func() {
		log.Printf("server running at: %s", api.Server.Addr)
		if err := api.Server.ListenAndServe(); err != nil {
			log.Fatalf("bye bye! %s\n", err.Error())
		}

	}()
}

func (api *Api) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	api.Server.Shutdown(ctx)
}

func WithRoutes(r []RouteMeta) optionFn {
	return func(o *option) {
		o.routes = append(o.routes, r...)
	}
}

type option struct {
	routes []RouteMeta
}

type optionFn = func(*option)

func New(options ...optionFn) (api *Api, err error) {
	mux := http.NewServeMux()

	o := &option{
		routes: make([]RouteMeta, 0),
	}

	for _, fn := range options {
		fn(o)
	}

	for _, r := range o.routes {
		mux.Handle(r.Path, r.Handler)
	}

	api = &Api{
		Server: &http.Server{
			Addr:        "localhost:1337",
			Handler:     mux,
			IdleTimeout: 30,
		},
	}

	return
}
