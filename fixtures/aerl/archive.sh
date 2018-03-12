#!/bin/bash

# PATHS
SRC="data"
DST="archive/honu-experiment-$(date +"%Y%m%d%H%M").tgz"

# Create the archive
tar -czf $DST $SRC

# Delete old files in the src directory
rm -rf $SRC/alia-*
rm -rf $SRC/run-*
