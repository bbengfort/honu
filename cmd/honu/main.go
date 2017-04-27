package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bbengfort/honu"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

func main() {

	// Load the .env file if it exists
	godotenv.Load()

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
					Name:   "a, addr",
					Usage:  "ip address to serve on",
					Value:  ":3264",
					EnvVar: "HONU_SERVER_ADDR",
				},
				cli.BoolFlag{
					Name:   "r, relax",
					Usage:  "relax to sequential consistency",
					EnvVar: "HONU_SEQUENTIAL_CONSISTENCY",
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
					Name:   "a, addr",
					Usage:  "ip address of the remote server",
					Value:  "localhost:3264",
					EnvVar: "HONU_SERVER_ADDR",
				},
				cli.StringFlag{
					Name:   "k, key",
					Usage:  "name or key to get the value for",
					EnvVar: "HONU_LOCAL_KEY",
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
					Name:   "a, addr",
					Usage:  "ip address of the remote server",
					Value:  "localhost:3264",
					EnvVar: "HONU_SERVER_ADDR",
				},
				cli.StringFlag{
					Name:   "k, key",
					Usage:  "name or key to get the value for",
					EnvVar: "HONU_LOCAL_KEY",
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
					Name:   "a, addr",
					Usage:  "ip address of the remote server",
					Value:  "localhost:3264",
					EnvVar: "HONU_SERVER_ADDR",
				},
				cli.StringFlag{
					Name:   "k, key",
					Usage:  "name or key to create a workload on",
					EnvVar: "HONU_LOCAL_KEY",
				},
				cli.StringFlag{
					Name:   "d, duration",
					Usage:  "parsable duration to run for",
					Value:  "10s",
					EnvVar: "HONU_RUN_DURATION",
				},
				cli.StringFlag{
					Name:  "w, results",
					Usage: "path on disk to write results to",
					Value: "",
				},
				cli.BoolFlag{
					Name:   "X, disabled",
					Usage:  "exit without running throughput",
					EnvVar: "HONU_RUN_DISABLED",
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
	server := honu.NewServer(c.Bool("relax"))

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
	if c.Bool("disabled") {
		fmt.Println("this client is disabled, exiting")
		return nil
	}

	duration, err := time.ParseDuration(c.String("duration"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	if err := client.Run(c.String("key"), duration, c.String("results")); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}
