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
