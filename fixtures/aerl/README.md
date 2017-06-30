# Reinforcement Learning of Anti-Entropy Topologies

This experiment tests automatic topology adaptation using reinforcement learning with anti-entropy. The idea is simple: anti-entropy selects a random peer in the network to synchronize with (reducing the entropy of consistency). Usually the random selection is uniform - that is all replicas have the same chance of being selected. However, what if the model learned what was most beneficial link was and selected it?

In this initial experiment we compare [epsilon greedy](https://en.wikipedia.org/wiki/Multi-armed_bandit#Semi-uniform_strategies) selection (semi-uniform) with random selection to see if the topology is modified beneficially. This is the start to a larger study, but shows the possibility for RL in topology selection.

## Topology

In this system we will run 12 replica servers, 4 processes on each of the following machines:

- nevis (College Park)
- hyperion (College Park)
- lagoon (College Park)

Each process will be run on a separate port per host: 3264-3267. All replicas will be associated to all other replica peers for anti-entropy selection. The anti-entropy delay will be set to 200ms.

For each sample two workloads will be generated, one on lagoon and one on hyperion. They will connect to a local replica and continuously write to the same key. Local throughput is far greater than the anti-entropy delay (generally about 10k messages per second or 1000 messages per millisecond) so there is guaranteed to be a synchronization from the associated replica on every anti-entropy session.

## Method

We will test two anti-entropy neighbor selection methods:

- uniform random selection
- epsilon-greedy with epsilon = 0.1

The other dimension that we'll explore is the amount of time to learn the topology, initially we'll look at 30 seconds, but run 5, 10, 15, 20, 25, and 30 second versions of the system as well to see how long it takes to normalize the topology. 
