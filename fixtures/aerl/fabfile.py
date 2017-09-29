# fabfile
# Fabric command definitions for running anti-entropy reinforcement learning.
#
# Author:   Benjamin Bengfort <benjamin@bengfort.com>
# Created:  Tue Jun 13 22:26:00 2017 -0400
#
# Copyright (C) 2017 Bengfort.com
# For license information, see LICENSE.txt
#
# ID: fabfile.py [] benjamin@bengfort.com $

"""
Fabric command definitions for running anti-entropy reinforcement learning.
"""

##########################################################################
## Imports
##########################################################################

import os
import re

from fabric.api import execute
from fabric.api import env, run, cd, parallel, get


##########################################################################
## Environment
##########################################################################

# Names
NEVIS = "nevis.cs.umd.edu"
HYPERION = "hyperion.cs.umd.edu"
LAGOON = "lagoon.cs.umd.edu"
ERIS = "eris.cs.umd.edu"
SEDNA = "keleher.duckdns.org"

# Paths
workspace = "/data/honu"
repo = "~/workspace/go/src/github.com/bbengfort/honu"

# Fabric Env
env.hosts = [NEVIS, HYPERION, LAGOON, ERIS, SEDNA]

# Fabric Env
env.colorize_errors = True
env.use_ssh_config = True
env.forward_agent = True


##########################################################################
## Task Helper Functions
##########################################################################

def pproc_command(commands):
    """
    Creates a pproc command from a list of command strings.
    """
    commands = " ".join([
        "\"{}\"".format(command) for command in commands
    ])
    return "pproc {}".format(commands)


##########################################################################
## Honu Commands
##########################################################################

@parallel
def update():
    """
    Update honu by pulling the repository and installing the command.
    """
    with cd(repo):
        run("git pull")
        run("godep restore")
        run("go install ./...")


@parallel
def version():
    """
    Get the current honu version number
    """
    with cd(repo):
        run("honu --version")
