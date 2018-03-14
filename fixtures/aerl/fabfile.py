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
import json
import pytz

from datetime import datetime
from StringIO import StringIO
from tabulate import tabulate
from operator import itemgetter
from collections import defaultdict
from dotenv import load_dotenv, find_dotenv
from dateutil.parser import parse as date_parse

from fabric.contrib import files
from fabric.api import env, run, cd, get
from fabric.colors import red, green, cyan
from fabric.api import parallel, task, runs_once, execute


##########################################################################
## Environment Helpers
##########################################################################

# Load the host information
def load_hosts(path):
    with open(path, 'r') as f:
        return json.load(f)


def load_host_regions(hosts):
    locations = defaultdict(list)
    for host in hosts:
        loc = " ".join(host.split("-")[1:-1])
        locations[loc].append(host)
    return locations


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
regions = load_host_regions(hosts)
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
    return " ".join(["--{} {}".format(k,v).strip() for k,v in args.items()])


def make_client_args(config, host):
    cmds = []
    name = addrs[host]
    if name not in config["clients"]["hosts"]:
        return cmds

    for conf in config['clients']['configs']:
        args = conf.copy()
        args['prefix'] = config["clients"]["hosts"][name]
        cmds.append(
            " ".join(["--{} {}".format(k,v).strip() for k,v in args.items()])
        )

    return cmds


##########################################################################
## Honu Commands
##########################################################################

@task
@parallel
def install():
    """
    Install epaxos for the first time on each machine
    """
    with cd(os.path.dirname(repo)):
        run("git clone git@github.com:bbengfort/honu.git")

    with cd(repo):
        run("godep restore")
        run("go install ./...")

    run("mkdir -p {}".format(workspace))


@task
@parallel
def uninstall():
    """
    Uninstall ePaxos on every machine
    """
    run("rm -rf {}".format(repo))
    run("rm -rf {}".format(workspace))


@task
@parallel
def update():
    """
    Update honu by pulling the repository and installing the command.
    """
    with cd(repo):
        run("git pull")
        run("godep restore")
        run("go install ./...")


@task
@parallel
def version():
    """
    Get the current honu version number
    """
    with cd(repo):
        run("honu --version")


@task
@parallel
def cleanup():
    """
    Cleans up results files so that the experiment can be run again.
    """
    names = (
        "metrics.json", "visibile_versions.log",
    )

    for name in names:
        path = os.path.join(workspace, name)
        run("rm -f {}".format(path))


@task
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
    for arg in args:
        command.append("honu bench {}".format(arg))

    with cd(workspace):
        run(pproc_command(command))


@task
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


@task
@parallel
def putkey(key="foo", value=None, visibility=True, geo="virginia", n_replicas=1):
    n_replicas = int(n_replicas)
    hostname = addrs[env.host]
    region = " ".join(hostname.split("-")[1:-1])

    # Perform geography filtering
    geo = geo.split(";")
    if region not in geo: return "ignoring {}".format(geo)
    if hostname not in regions[region][:n_replicas]: return

    if value is None:
        now = pytz.timezone('America/New_York').localize(datetime.now())
        now = now.strftime("%Y-%m-%d %H:%M:%S %z")
        value = '"created on {} at {}"'.format(hostname, now)

    cmd = "honu put -k {} -v {}".format(key, value)
    if visibility:
        cmd += " --visibility"

    run(cmd)


@parallel
def _getkey(key):
    return run("honu get -k {}".format(key), quiet=True)


@task
@runs_once
def getkey(key="foo"):
    """
    Returns the latest version of the key on the hosts specified
    """
    row = re.compile(r'version ([\d\.]+), value: (.*)', re.I)
    data = execute(_getkey, key)
    table = []

    for host, line in data.items():
        match = row.match(line)
        if match is None:
            version = 0.0
            value = red(line.split("\n")[-1])
        else:
            version, value = match.groups()
        table.append([host, float(version), value])

    table = [["Host", "Version", "Value"]] + table
    print(tabulate(table, tablefmt='simple', headers='firstrow', floatfmt=".2f"))



@parallel
def fetch_visibility():
    """
    Fetch and parse the visibility data
    """
    fd = StringIO()
    remote = os.path.join(workspace, "visibile_versions.log")
    get(remote, fd)
    rows = fd.getvalue().split("\n")
    return list(map(json.loads, filter(None, rows)))


@task
@runs_once
def visibility():
    """
    Print a table of the current key visibility listed
    """
    def loc_count(replicas):
        locs = defaultdict(int)
        for replica in replicas:
            loc = " ".join(replica.split("-")[1:-1])
            locs[loc] += 1
        output = [
            "{}-{}".format(l, c) for l, c in
            sorted(locs.items(), key=itemgetter(1), reverse=True)
        ]

        for idx in range(0, len(output), 2):
            output[idx] = cyan(output[idx])

        return " ".join(output)


    data = execute(fetch_visibility)
    n_hosts = len(data)
    versions = defaultdict(list)

    for host, vals in data.items():
        for val in vals:
            vers = "{Key} {Version}".format(**val)
            versions[vers].append((host, val['Timestamp']))

    table = [['Version', 'R', '%', 'L (secs)', 'Created', 'Updated', 'Regions']]
    for vers, timestamps in sorted(versions.items(), key=itemgetter(0)):
        replicas = [h[0] for h in timestamps]
        timestamps = [date_parse(h[1]) for h in timestamps]
        replicated = len(set(replicas))
        visibility = (float(replicated) / float(n_hosts)) * 100.0
        created = min(timestamps)
        updated = max(timestamps)
        latency = (updated - created).total_seconds()
        table.append([
            vers,
            replicated,
            "{:0.2f}".format(visibility),
            "{:0.2f}".format(latency),
            created.strftime("%Y-%m-%d %H:%M:%S"),
            updated.strftime("%Y-%m-%d %H:%M:%S"),
            loc_count(replicas)
        ])


    print(tabulate(table, tablefmt='simple', headers='firstrow'))
