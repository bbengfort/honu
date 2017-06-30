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

from copy import copy
from multiprocessing import Lock
from fabric.api import env, run, cd, parallel, get
from fabric.api import roles, task, execute, settings


##########################################################################
## Environment
##########################################################################

# Names
NEVIS = "nevis.cs.umd.edu"
HYPERION = "hyperion.cs.umd.edu"
LAGOON = "lagoon.cs.umd.edu"
CLIENTA = "client.hyperion.cs.umd.edu"
CLIENTB = "client.lagoon.cs.umd.edu"

# Paths
workspace = "/data/honu"

# Regular Expressions
ports = re.compile(r'.+\.(\d{4})')

# Fabric Env
env.colorize_errors = True
env.dedupe_hosts = False
env.hosts = addrs + [CLIENTA, CLIENTB]
env.user = "benjamin"


##########################################################################
## Helper Functions
##########################################################################

def unfix(s, prefix=None, suffix=None):
    """
    Remove a prefix or a suffix or both from a string.
    """
    if prefix and s.startswith(prefix):
        s = s[len(prefix):]

    if suffix and s.endswith(suffix):
        s = s[:-1 * len(suffix)]

    return s


def get_peers(addr):
    peers = set(addrs)
    peers.remove(addr)
    return ",".join(peers)


##########################################################################
## Honu Commands
##########################################################################

def serve(uptime="35s", addr=":3264"):
    peers = get_peers(addr)
    addr = ":" + ports.match(addr).group(1)

    with cd(workspace):
        cmd = "honu serve -d 200ms -a {} -u {} -w replica.jsonl -p {}".format(
            addr, uptime, peers
        )
        run(cmd)


def workload(duration="30s"):
    with cd(workspace):
        cmd = "honu run -w /dev/null -A -a :3264 -d {} -k foo".format(duration)
        run(cmd)


@parallel
def experiment():
    if env.host.startswith("client."):
        host = unfix(env.host, prefix="client.")
        execute(workload, host=host)

    else:
        addr = env.host_string
        host = unfix(addr, suffix=".3264")
        host = unfix(addr, suffix=".3265")
        host = unfix(addr, suffix=".3266")
        host = unfix(addr, suffix=".3267")
        execute(serve, host=host, addr=addr)
