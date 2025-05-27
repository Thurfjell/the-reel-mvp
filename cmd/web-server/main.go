package main

import (
	"flag"
	"log"
	"movez/internal/api"
	"movez/internal/api/movie"
	"movez/internal/api/movie/http"
	"movez/internal/api/movie/sql33t"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	accessTokenFlag := flag.String("token", "", "themoviedb access_token")
	sqliteNameFlag := flag.String("sqlite_path", "testar", "Path to your sqlite db")

	flag.Parse()
	if len(*accessTokenFlag) == 0 {
		flag.Usage()
		log.Fatalln("no token flag")
	}

	movieService := http.NewSevice(*accessTokenFlag)
	if err := sql33t.Up(*sqliteNameFlag); err != nil {
		log.Fatalf("ruh roh... %+v", err)
	}

	movieCache, err := sql33t.NewCache(*sqliteNameFlag)
	if err != nil {
		log.Fatalf("ruh roh.. %+v", err)
	}

	movieHandler, err := movie.NewHandler(movie.NewCachedMovieService(movieService, movieCache))
	if err != nil {
		log.Fatalf("ruh roh.. %+v", err)
	}

	apiServer, err := api.New(api.WithRoutes(movieHandler.Routes()))

	if err != nil {
		log.Fatalf("ruh roh.. %+v", err)
	}

	apiServer.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	movieCache.Close()
	apiServer.Close()
}
