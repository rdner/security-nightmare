package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/pragmader/security-nightmare/server"
)

const (
	address         = ":666"
	attackerAddress = ":777"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s, err := server.NewServer(ctx, "./feeds.db")
	if err != nil {
		log.Fatal(err)
	}

	attacker := createAttackerServer()
	serv := createServer(s)
	done := listenSignals(s, attacker, serv)

	log.Printf("Starting the attacker server on %s", attackerAddress)
	go attacker.ListenAndServe()

	log.Printf("Starting the server on %s", address)
	serv.ListenAndServe()

	<-done
}

func createServer(s server.Server) *http.Server {
	r := chi.NewRouter()
	r.Get("/feed", s.Feed)
	r.Get("/client-side", s.ClientSideFeed)
	r.Post("/feed", s.Add)
	r.Delete("/feed", s.Delete)
	r.Get("/feed/delete", s.Delete)

	return &http.Server{Addr: address, Handler: r}
}

func createAttackerServer() *http.Server {
	a := server.NewAttacker()
	r := chi.NewRouter()
	r.Get("/opener", a.OpenerCase)
	r.Get("/csrf", a.CSRFCase)
	r.Get("/csrf_form", a.CSRFFormCase)
	r.Get("/gotcha", a.Gotcha)

	return &http.Server{Addr: attackerAddress, Handler: r}
}

func listenSignals(s server.Server, httpServers ...*http.Server) <-chan bool {
	done := make(chan bool)
	signalChan := make(chan os.Signal) // listen to interrupt signals
	signal.Notify(
		signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT,
	)

	go func() {
		<-signalChan
		log.Print("Signal received, shutting down...")
		s.Shutdown()
		for _, server := range httpServers {
			server.Shutdown(context.Background())
		}
		log.Print("Finished")
		done <- true
	}()
	return done
}
