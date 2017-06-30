# Hierarchical Consensus

- compare stand alone Raft (all replicas in the quorum) to HC with subquorum sizes of 3 and 5
- x axis: number of replicas (n) in the system, (3, 5, 6, 9, 10, 12, 15, 18, 20, 21, 24, 25)
- y axis: throughput (writes/second)

Setup:

1. start n replicas, wait for them to converge on a leader (initially, HC is pre-configured into tiers)
2. start n workload generators, one local to each replica, each writing to a different key (no conflicts)
3. respond to client on commit
4. each workload runs for 1m then shuts down


Goal: show that HC has a higher throughput than Raft thanks to the load balancing of the subquorums
