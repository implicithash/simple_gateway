package app

import (
	"context"
	"fmt"
	"github.com/implicithash/simple_gateway/src/handlers"
	"github.com/implicithash/simple_gateway/src/services"
	"github.com/implicithash/simple_gateway/src/utils/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// StartApplication starts an application
func StartApplication() {
	killSignalChan := getKillSignalChan()
	srv := startServer(":8000")

	waitForKillSignal(killSignalChan)

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	handlers.StopPool()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println(err)
	}
	os.Exit(0)
}

func startServer(serverURL string) *http.Server {
	if err := config.Setup(); err != nil {
		return nil
	}
	handlers.RunPool()
	router := handlers.MapUrls()
	srv := &http.Server{
		Addr:         serverURL,
		WriteTimeout: 500 * time.Millisecond,
		ReadTimeout:  5 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      router,
	}

	go func() {
		fmt.Println("Starting server at :8000")
		log.Fatal(srv.ListenAndServe())
	}()

	return srv
}

func getKillSignalChan() chan os.Signal {
	osKillSignalChan := make(chan os.Signal, 1)
	signal.Notify(osKillSignalChan)
	return osKillSignalChan
}

func waitForKillSignal(killSignalChan <-chan os.Signal) {
	killSignal := <-killSignalChan
	switch killSignal {
	case os.Interrupt:
		log.Println("got SIGINT...")
		log.Println(fmt.Sprintf("Statistics: incoming: %d, outgoing: %d", services.Limiter.IncomingCounter, services.Limiter.OutgoingCounter))
		log.Println("Stopping server...")
	case syscall.SIGTERM:
		log.Println("got SIGTERM...")
	}
}
