# Honu

**Throughput testing for a simple, volatile, in-memory key/value store.**

## Modes

Currently Honu can implement different modes of operation for a variety of experimental circumstances. The current dimensions are as follows:

**Consistency**

- linearizable: the entire store is locked on writes and read-locked on reads
- sequential: only the object for the specified keys is locked on accesses

**Replication**

- standalone: no replication, the server exists in isolation
- anti-entropy: full replication of objects using bilateral anti-entropy

## Getting Started

First, fetch the code and install it:

    $ go get github.com/bbengfort/honu/...

You should now have `honu` on your `$PATH`:

    $ honu --help

### Servers

You can run a server as follows:

    $ honu serve

Which by default will run on `:3264`, you can specify a different address with the `-a`, `--addr` flag or the `$HONU_SERVER_ADDR` environment variable.

The default consistency mode for a server is linearizable, this means that the  entire store is read locked to Get a value and write locked to Put a value. All versions have a single monotonically increasing version number and accesses between objects are totally ordered.

> NOTE: linearizable here is only with respect to the local replica. Forks can still occur if concurrent accesses happen before replication.

In order to switch to sequential mode (each object is accessed independently with respect to key), specify the `-r`, `--relax` or set the `$HONU_SEQUENTIAL_CONSISTENCY` environment variable to true.

### Clients and Throughput

Clients can be run as follows:

    $ honu put -k foo -v bar
    $ honu get -k foo

By default, the client will connect to a local server or the one specified by the `$HONU_SERVER_ADDR`; to specify a different address to connect to, use the `-a`, `--addr` flag.

The throughput experiment can be run for a specified duration as follows:

    $ honu run -d 30s -k foo

This will test how many writes to the server can occur within 30 seconds.

### Version History

The server can be quit using `CTRL+C`, it will perform any clean up required and shutdown. If you would like to dump a version log from the server on shutdown, run the server with the `-o`, `--objects` option:

    $ honu server --objects path/to/version.log

This will write out the view of the replica; that is the version history that the replica has seen to a JSON file locally. Note that the version history is the chain or tree of versions that have been applied to objects, not the actual values!

### Replication

For replication, servers need to know their peers. This can be specified with a comma delimited list using the `-p`, `--peers` flag, or using the `$HONU_PEERS` environment variable. Replication is the default mode, but will not occur if there are no peers (e.g. an empty string) or if the `-s`, `--standalone` flag is set (alternatively the `$HONU_STANDALONE_MODE` environment variable is set to true).

Replication is currently implemented by bilateral anti-entropy. Specify the anti-entropy delay with the `-d`, `--delay` flag or the `$HONU_ANTI_ENTROPY_DELAY` environment variable. This value must be a parseable duration, the default is `1s`.

## Configuration

You can create a .env file in the local directory that you're running honu from (or export environment variables) with the following configuration:

```
# Replica Configuration
HONU_PROCESS_ID=1
HONU_SERVER_ADDR=:3264
HONU_PEERS=""
HONU_STANDALONE_MODE=false
HONU_ANTI_ENTROPY_DELAY=1s
HONU_BANDIT_STRATEGY=uniform
HONU_SEQUENTIAL_CONSISTENCY=false
HONU_RANDOM_SEED=42

# Client and Experiment Configuration
HONU_LOCAL_KEY=foo
HONU_RUN_DURATION=30s
HONU_RUN_DISABLED=false
```

Hopefully this will help run experiments without typing in tons of arguments at the command line.

## Experiments and Results

For more, see [experiments and fixtures](fixtures/README.md), which contains detailed listings of the various results and methods that employ Honu.
