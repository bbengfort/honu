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
from dotenv import load_dotenv, find_dotenv
from fabric.api import env, run, cd, parallel, get


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
## Environment
##########################################################################

## Load the environment
load_dotenv(find_dotenv())

## Local paths
fixtures = os.path.dirname(__file__)
hostinfo = os.path.join(fixtures, "hosts.json")

## Remote Paths
workspace = "/data/honu"
repo = "~/workspace/go/src/github.com/bbengfort/honu"

## Load Hosts
hosts = load_hosts(hostinfo)
addrs = {info['hostname']: host for host, info in hosts.items()}
env.hosts = sorted(list(hosts.keys()))

## Fabric Env
env.user = "ubuntu"
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


def make_replica_args(config, host):
    name = addrs[host]
    info = hosts[name]

    if name not in config["replicas"]["hosts"]:
        return None

    args = config['replicas']['config'].copy()
    args['pid'] = int(name.split("-")[-1])

    args['peers'] = ",".join([
        hosts[peer]['hostname'] + ":3264"
        for peer in config["replicas"]["hosts"]
        if peer != name
    ])
    return " ".join(["--{} {}".format(k,v) for k,v in args.items()])


def make_client_args(config, host):
    name = addrs[host]
    if name not in config["clients"]["hosts"]:
        return None

    args = config['clients']['config'].copy()
    args['prefix'] = config["clients"]["hosts"][name]
    return " ".join(["--{} {}".format(k,v) for k,v in args.items()])


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

    # Create the serve command
    args = make_replica_args(config, env.host)
    if args:
        command.append("honu serve {}".format(args))

    # Create the bench command
    args = make_client_args(config, env.host)
    if args:
        command.append("honu bench {}".format(args))

    with cd(workspace):
        run(pproc_command(command))


@parallel
def getmerge(name="metrics.json", path="data", suffix=None):
    """
    Get the results.json and the metrics.json files and save them with the
    specified suffix to the localpath.
    """
    remote = os.path.join(workspace, name)
    hostname = addrs[env.host]
    local = os.path.join(path, hostname, add_suffix(name, suffix))
    local  = unique_name(local)
    if files.exists(remote):
        get(remote, local)
