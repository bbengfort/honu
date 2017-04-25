# Honu

**Throughput testing for a simple, volatile, in-memory key/value store.**

## Getting Started

First, fetch the code and install it:

    $ go get github.com/bbengfort/honu/...

You should now have `honu` on your `$PATH`:

    $ honu --help

You can run the server as follows:

    $ honu serve

Which by default will run on `:3264`, you can specify a different address with the `-a` flag. Clients can be run as follows:

    $ honu put -k foo -v bar
    $ honu get -k foo

The throughput experiment can be run for a specified duration as follows:

    $ honu run -d 30s -k foo

This will test how many writes to the server can occur within 30 seconds.

## Configuration

You can create a .env file in the local directory that you're running honu from (or export environment variables) with the following configuration:

```
HONU_SERVER_ADDR=192.168.35.1:3264
HONU_LOCAL_KEY=foo
HONU_RUN_DURATION=30s
```

Hopefully this will help run experiments without typing in tons of arguments at the command line.
