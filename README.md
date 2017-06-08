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
- raft: replication using broadcast consensus (not implemented)

## Getting Started

First, fetch the code and install it:

    $ go get github.com/bbengfort/honu/...

You should now have `honu` on your `$PATH`:

    $ honu --help

### Servers

You can run a server as follows:

    $ honu serve

Which by default will run on `:3264`, you can specify a different address with the `-a` flag or the `$HONU_SERVER_ADDR` environment variable.

For replication, servers need to know their peers. This can be specified with a semi-colon delimited list using the `-peers` flag, or using the `$HONU_PEERS` environment variable.

### Clients and Throughput

Clients can be run as follows:

    $ honu put -k foo -v bar
    $ honu get -k foo

By default, the client will connect to a local server or the one specified by the `$HONU_SERVER_ADDR`; to specify a different address to connect to, use the `-a` flag.

The throughput experiment can be run for a specified duration as follows:

    $ honu run -d 30s -k foo

This will test how many writes to the server can occur within 30 seconds.

### Version History

The server can be quit using `CTRL+C`, it will perform any clean up required and shutdown. If you would like to dump a version log from the server on shutdown, run the server with the `-objects` option:

    $ honu server -objects

This will write out the view of the replica; that is the version history that the replica has seen to a JSON file locally.

## Configuration

You can create a .env file in the local directory that you're running honu from (or export environment variables) with the following configuration:

```
# Replica Configuration
HONU_PROCESS_ID=1
HONU_SERVER_ADDR=:3264
HONU_PEERS=""
HONU_SEQUENTIAL_CONSISTENCY=false

# Client and Experiment Configuration
HONU_LOCAL_KEY=foo
HONU_RUN_DURATION=30s
HONU_RUN_DISABLED=false
```

Hopefully this will help run experiments without typing in tons of arguments at the command line.

## Experiments and Results

For more, see [experiments and fixtures](fixtures/README.md), which contains detailed listings of the various results and methods that employ Honu.
