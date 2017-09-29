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
import json

from fabric.contrib import files
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
## Environment Helpers
##########################################################################

# Load the host information
def load_hosts(path):
    with open(path, 'r') as f:
        return json.load(f)


def parse_bool(val):
    if isinstance(val, basestring):
        val = val.lower().strip()
        if val in {'yes', 'y', 'true', 't', '1'}:
            return True
        if val in {'no', 'n', 'false', 'f', '0'}:
            return False
    return bool(val)


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


def round_robin(n, host, hosts=env.hosts):
    """
    Returns a number n (of clients) for the specified host, by allocating the
    n clients evenly in a round robin fashion. For example, if hosts = 3 and
    n = 5; then this function returns 2 for host[0], 2 for host[1] and 1 for
    host[2].
    """
    num = n / len(hosts)
    idx = hosts.index(host)
    if n % len(hosts) > 0 and idx < (n % len(hosts)):
        num += 1
    return num


def add_suffix(path, suffix=None):
    if suffix:
        base, ext = os.path.splitext(path)
        path = "{}-{}{}".format(base, suffix, ext)
    return path


def unique_name(path, start=0, maxtries=1000):
    for idx in range(start+1, start+maxtries):
        ipath = add_suffix(path, idx)
        if not os.path.exists(ipath):
            return ipath

    raise ValueError(
        "could not get a unique path after {} tries".format(maxtries)
    )

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


@parallel
def cleanup():
    """
    Cleans up results files so that the experiment can be run again.
    """
    names = ("results", "metrics")
    exts = (".json", ".jsonl")

    for name in names:
        for ext in exts:
            path = os.path.join(workspace, name+ext)
            run("rm -f {}".format(path))


@parallel
def bench(config="config.json"):
    """
    Run all servers on the host as well as benchmarks for the number of
    clients specified.
    """
    command = []

    # load the configuration
    with open(config, 'r') as f:
        config = json.load(f)

    peers = [
        replica["addr"]
        for host in config.values()
        for replica in host["replicas"]
    ]

    for proc in config[env.host]['replicas']:
        proc["peers"] = ",".join([peer for peer in peers if peer != proc["addr"]])
        proc["addr"] = ":" + proc["addr"].split(":")[-1]
        
        args = " ".join(["--{} {}".format(k,v) for k,v in proc.items()])
        cmd = "honu serve {}".format(args)
        command.append(cmd)

    for client in config[env.host]['clients']:
        args = " ".join(["--{} {}".format(k,v) for k,v in client.items()])
        cmd = "honu bench {}".format(args)
        command.append(cmd)

    with cd(workspace):
        run(pproc_command(command))


@parallel
def getmerge(name="metrics.json", path="data", suffix=None):
    """
    Get the results.json and the metrics.json files and save them with the
    specified suffix to the localpath.
    """
    remote = os.path.join("/", "data", "honu", name)
    local = os.path.join(path, env.host, add_suffix(name, suffix))
    local  = unique_name(local)
    if files.exists(remote):
        get(remote, local)
