package honu

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc/grpclog"
)

//===========================================================================
// Package Initialization
//===========================================================================

// PackageVersion of the current Honu implementation
const PackageVersion = "0.7"

// Initialize the package and random numbers, etc.
func init() {
	// Set the random seed to something different each time.
	rand.Seed(time.Now().Unix())

	// Initialize our debug logging with our prefix
	logger = log.New(os.Stdout, "[honu] ", log.Lmicroseconds)

	// Stop the grpc verbose logging
	grpclog.SetLogger(noplog)
}

//===========================================================================
// OS Signal Handlers
//===========================================================================

func signalHandler(shutdown func() error) {
	// Make signal channel and register notifiers for Interupt and Terminate
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)

	// Block until we receive a signal on the channel
	<-sigchan

	// Defer the clean exit until the end of the function
	defer os.Exit(0)

	// Shutdown now that we've received the signal
	debug("shutting down the honu server or client")
	if err := shutdown(); err != nil {
		warn("could not gracefully shutdown: %s", err)
		os.Exit(1)
	}

	// Declare graceful shutdown.
	info("honu has gracefully shutdown")
}
