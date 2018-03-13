#!/bin/bash
# Run the anti-entropy experiments

# Describe the time format
TIMEFORMAT="experiment completed in %2lR"

# Setup variables and paths
DATA="data"
HOSTS="hosts.json"
CONFIGS="$DATA/config*.json"

# Determine the number of runs
RUNS=$1
if [ -z "$RUNS" ]; then
    echo "specify the number of times to run the experiment"
    exit
fi

# Execute all of the experiments
time {
    # Ensure package is up to date
    fab update
    fab version

    # Perform housekeeping tasks
    fab cleanup
    cp $HOSTS $DATA

    # Conduct the experiment $RUNS times
    for (( I=0; I<$RUNS; I+=1 )); do
        RESULTS="$DATA/run-$I"
        mkdir $RESULTS

        # Run each experiment whose parameters are defined in a config
        for CONFIG in $CONFIGS; do
            fab bench:config=$CONFIG
            fab getmerge:path=$RESULTS
            fab getmerge:name=visibile_versions.log,path=$RESULTS
            fab cleanup
        done
    done
}
