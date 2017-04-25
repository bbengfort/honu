package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bbengfort/honu"
	"github.com/urfave/cli"
)

func main() {

	// Instantiate the command line application
	app := cli.NewApp()
	app.Name = "honu"
	app.Version = "0.1"
	app.Usage = "throughput testing for a volatile, in-memory key/value store"

	// Define commands available to the application
	app.Commands = []cli.Command{
		{
			Name:     "serve",
			Usage:    "run the honu key/value storage server",
			Action:   serve,
			Category: "server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, addr",
					Usage: "ip address to serve on",
					Value: ":3264",
				},
			},
		},
		{
			Name:     "get",
			Usage:    "get the current value and version associated with a key",
			Action:   get,
			Category: "client",
			Before:   initClient,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, addr",
					Usage: "ip address of the remote server",
					Value: "localhost:3264",
				},
				cli.StringFlag{
					Name:  "k, key",
					Usage: "name or key to get the value for",
				},
			},
		},
		{
			Name:     "put",
			Usage:    "put a value associated with a key",
			Action:   put,
			Category: "client",
			Before:   initClient,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, addr",
					Usage: "ip address of the remote server",
					Value: "localhost:3264",
				},
				cli.StringFlag{
					Name:  "k, key",
					Usage: "name or key to get the value for",
				},
				cli.StringFlag{
					Name:  "v, value",
					Usage: "value to write to the storage server",
				},
			},
		},
		{
			Name:     "run",
			Usage:    "run the throughput experiment",
			Action:   run,
			Category: "client",
			Before:   initClient,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, addr",
					Usage: "ip address of the remote server",
					Value: "localhost:3264",
				},
				cli.StringFlag{
					Name:  "k, key",
					Usage: "name or key to create a workload on",
					Value: "",
				},
				cli.StringFlag{
					Name:  "d, duration",
					Usage: "parsable duration to run for",
					Value: "10s",
				},
				cli.StringFlag{
					Name:  "w, results",
					Usage: "path on disk to write results to",
					Value: "",
				},
			},
		},
	}

	// Run the CLI program
	app.Run(os.Args)
}

//===========================================================================
// Server Commands
//===========================================================================

// Run the storage server
func serve(c *cli.Context) error {
	server := honu.NewServer()

	if err := server.Run(c.String("addr")); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

//===========================================================================
// Client Commands
//===========================================================================

var client *honu.Client

// Initialize the client
func initClient(c *cli.Context) error {
	client = new(honu.Client)
	if err := client.Connect(c.String("addr")); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

// Get a value for a key
func get(c *cli.Context) error {
	value, version, err := client.Get(c.String("key"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	fmt.Printf("version %d, value: %s\n", version, string(value))
	return nil
}

// Put a value for a key
func put(c *cli.Context) error {
	version, err := client.Put(c.String("key"), []byte(c.String("value")))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	fmt.Printf("key %s now at version %d\n", c.String("key"), version)
	return nil
}

// Run the throughput experiment
func run(c *cli.Context) error {
	duration, err := time.ParseDuration(c.String("duration"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	if err := client.Run(c.String("key"), duration, c.String("results")); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}
