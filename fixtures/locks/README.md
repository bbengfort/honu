# Comparison of Standalone Storage Locking Mechanisms

This experiment analyzes the effectiveness of locking mechanism for a standalone (non-replicated) store that has multiple concurrent processes writing to it. In particular it is a comparison of the following strategies:

1. Linearizable: the entire data structure is locked during accesses.
2. Sequential: the object being accessed is locked during accesses.

Note that there are read locks (on Get) and write locks (on Put). A read lock allows multiple readers but blocks until a write lock is complete. Write locks occur one at a time. Locking is required because the gRPC implementation uses go routines to handle incoming RPC requests, therefore multiple concurrent accesses to the data store are possible.

The data structure in question is a map of key -> entry pairs where keys are strings and an entry is a struct. The linearizable store locks the entire map while it is accessing it, so that keys are read and written in order and in sequence to each other. To show how this works in practice, only a single monotonically increasing version sequence is used across all objects.

Sequential storage on the other hand initiates a read lock to fetch a given entry, then adds the appropriate lock (read or write) to the entry, releasing the global lock. A global write lock is only used when adding a new entry to the store.

Because sequential storage has more locks/unlocks it should be slower for accesses of multiple threads to a single key than linearizable storage. However, for accesses of multiple threads to multiple keys, it should be much faster than linearizable.

## Method

Using the metal servers: nevis, hyperion, and lagoon I ran a standalone server on nevis with both linearizable and sequential versions of the storage; as well as an increasing number of clients on hyperion and lagoon. There were two modes of clients, accessing a single key and each accessing their own key.

Therefore the experimental dimensions are as follows:

1. sequential vs. linearizable
2. multi-key vs. single-key

Resulting in 4 total experiments. Each experiment has several trials with an increasing number of clients from 1-32. Clients are distributed to hyperion and lagoon in a round-robin fashion such that at most only 16 processes on each box are running at a time. At conclusion we had 128 data points, 32 in each of our four experiments.

## Results
