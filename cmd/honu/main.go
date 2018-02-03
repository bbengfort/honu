package main

import (
	"fmt"
	"os"
	"strings"
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
	app.Version = honu.PackageVersion
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
					Value:  honu.DefaultAddr,
					EnvVar: "HONU_SERVER_ADDR",
				},
				cli.BoolFlag{
					Name:   "r, relax",
					Usage:  "relax to sequential consistency",
					EnvVar: "HONU_SEQUENTIAL_CONSISTENCY",
				},
				cli.Uint64Flag{
					Name:   "i, pid",
					Usage:  "unique process id of server",
					Value:  1,
					EnvVar: "HONU_PROCESS_ID",
				},
				cli.StringFlag{
					Name:   "p, peers",
					Usage:  "comma delmited list of address of remote replicas",
					EnvVar: "HONU_PEERS",
				},
				cli.StringFlag{
					Name:   "d, delay",
					Usage:  "parsable duration of anti-entropy delay",
					Value:  "1s",
					EnvVar: "HONU_ANTI_ENTROPY_DELAY",
				},
				cli.BoolFlag{
					Name:   "s, standalone",
					Usage:  "disable replication and run in standalone mode",
					EnvVar: "HONU_STANDALONE_MODE",
				},
				cli.StringFlag{
					Name:   "u, uptime",
					Usage:  "pass a parsable duration to shut the server down after",
					EnvVar: "HONU_SERVER_UPTIME",
				},
				cli.StringFlag{
					Name:   "w, stats",
					Usage:  "path on disk to write JSON stats to on shutdown",
					Value:  "",
					EnvVar: "HONU_SERVER_RESULTS",
				},
				cli.StringFlag{
					Name:  "c, history",
					Usage: "path on disk to write version history to on shutdown",
					Value: "",
				},
				cli.StringFlag{
					Name:   "b, bandit",
					Usage:  "bandit strategy for random peer selection",
					Value:  "uniform",
					EnvVar: "HONU_BANDIT_STRATEGY",
				},
				cli.Float64Flag{
					Name:   "e, epsilon",
					Usage:  "value of epsilon for epsilon greedy selection",
					Value:  0.2,
					EnvVar: "HONU_BANDIT_EPSILON",
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
					Value:  "localhost" + honu.DefaultAddr,
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
					Value:  "localhost" + honu.DefaultAddr,
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
			Name:     "bench",
			Usage:    "run the throughput experiment",
			Action:   bench,
			Category: "client",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "a, addr",
					Usage:  "ip address of the remote server",
					Value:  "localhost" + honu.DefaultAddr,
					EnvVar: "HONU_SERVER_ADDR",
				},
				cli.StringFlag{
					Name:   "d, duration",
					Usage:  "parsable duration to run for",
					Value:  "10s",
					EnvVar: "HONU_RUN_DURATION",
				},
				cli.StringFlag{
					Name:  "D, delay",
					Usage: "parseable duration to delay start of benchmark",
					Value: "",
				},
				cli.StringFlag{
					Name:   "o, results",
					Usage:  "path on disk to write results to",
					Value:  "",
					EnvVar: "HONU_CLIENT_RESULTS",
				},
				cli.IntFlag{
					Name:  "w, workers",
					Usage: "number of worker clients to initialize",
					Value: 1,
				},
				cli.StringFlag{
					Name:  "p, prefix",
					Usage: "key for clients to access, char for prefix, blank for random",
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
	// Create the server
	server := honu.NewServer(c.Uint64("pid"), c.Bool("relax"))

	// Parse the peers variable
	var peers []string
	if c.String("peers") != "" {
		peers = strings.Split(c.String("peers"), ",")
	}

	// Set the stats and version dump paths
	server.Measure(c.String("stats"), c.String("history"))

	// Run replication service
	if !c.Bool("standalone") && len(peers) > 0 {
		// Parse the delay variable
		delay, err := time.ParseDuration(c.String("delay"))
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		bandit := c.String("bandit")
		epsilon := c.Float64("epsilon")

		if err := server.Replicate(peers, delay, bandit, epsilon); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
	}

	// Set the uptime timer
	if c.String("uptime") != "" {
		// Parse the delay variable
		uptime, err := time.ParseDuration(c.String("uptime"))
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		server.Uptime(uptime)
	}

	// Run the server (blocks)
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

	fmt.Printf("version %s, value: %s\n", version, string(value))
	return nil
}

// Put a value for a key
func put(c *cli.Context) error {
	version, err := client.Put(c.String("key"), []byte(c.String("value")))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	fmt.Printf("key %s now at version %s\n", c.String("key"), version)
	return nil
}

// Run the throughput experiment
func bench(c *cli.Context) error {
	duration, err := time.ParseDuration(c.String("duration"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	// If a delay is specified parse how long to delay for
	var delay time.Duration
	if c.String("delay") != "" {
		if delay, err = time.ParseDuration(c.String("delay")); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
	}

	extra := make(map[string]interface{})
	bench, err := honu.NewBenchmark(c.Int("workers"), c.String("prefix"), extra)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	if err := bench.Run(c.String("addr"), c.String("results"), duration, delay); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}
