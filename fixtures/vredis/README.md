# Honu vs. Redis 

Running the redis v4.0.2 benchmark for PUT:

```
$ redis-benchmark -t set -n 100000 -q
SET: 86355.79 requests per second
```

Running Honu v0.6 benchmark with no modifications:

```
$ honu serve -s 
$ honu bench -w 50 
366805 accesses (0 reads, 366805 writes) in 10.004744107s -- 36663.1066 accesses/second
```

And with "relaxed consistency" (using SequentialStore):

```
$ honu serve -s -r
$ honu bench -w 50 
378066 accesses (0 reads, 378066 writes) in 10.007684404s -- 37777.5702 accesses/second
```

So Redis is 2.3x faster than Honu in its current form. 

## Notes 

Benchmarks run on MacBook Pro (15-inch, 2017) macOS High Sierra Version 10.13.2. 
Processor 3.1 GHz Intel Core i7; 16GB memory. 
