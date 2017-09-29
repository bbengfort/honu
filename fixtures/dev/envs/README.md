# Environments

Use the environments in this folder to export specific environment variables
to influence how a local development cluster runs. For example:

```
$ source envs/alpha.sh
$ source envs/honu.sh
```

Note that the order is important. These scripts will add the following environment variables to your shell:

- `$NAME`: the name of the replica to run, e.g. `"alpha"` in this case.
- `$SEED`: the random seed of the replica
- `$HONU`: the location of the `main.go` file
- `$SERVE`: a command to `go run $HONU serve -n $NAME -s $SEED`
- `$BENCH`: a command to `go run $HONU bench`

This should make it fairly easy to run local clusters and experiments.
