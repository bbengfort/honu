#!/bin/bash

SRC=data/
DST="archive/honu-experiment-$(date +"%Y%m%d%H%M").tgz"
tar -czvf $DST $SRC
