#!/bin/bash 
# Run the anti-entropy experiments 

# Describe the time format
TIMEFORMAT="experiment completed in %2lR"

time {
    # Step One: Ensure the package is up to date 
    fab update 
    fab version 

    # Step Two: Clean out any old results that linger 
    fab cleanup 

    # Step Three: Execute each configuration file 
    for (( I=1; I<=5; I+=1 )); do 
        fab bench:config=data/config-$I.json
        fab getmerge 
        fab cleanup 
    done 
}
