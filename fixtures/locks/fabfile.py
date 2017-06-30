# fabfile
# Fabric command definitions for running lock tests.
#
# Author:   Benjamin Bengfort <benjamin@bengfort.com>
# Created:  Tue Jun 13 12:47:15 2017 -0400
#
# Copyright (C) 2017 Bengfort.com
# For license information, see LICENSE.txt
#
# ID: fabfile.py [] benjamin@bengfort.com $

"""
Fabric command definitions for running lock tests.
"""

##########################################################################
## Imports
##########################################################################

import os
import random

from fabric.api import env, run, cd, parallel, get
from fabric.api import roles, task, execute, settings


class KeyGen(object):

    def __init__(self, n=3):
        self.n = n
        self.conson = "BCDFGHJKLMNPQRSTVWXZ"
        self.vowels = "AEIOUY"
        self.keys   = set([])

    def generate(self):
        word = ""
        for idx in range(self.n):
            if idx % 2 == 0:
                word += random.choice(self.conson)
            else:
                word += random.choice(self.vowels)

        if word in self.keys:
            return self.generate()

        self.keys.add(word)
        return word

    def __call__(self):
        return self.generate()


def strpbool(arg):
    if arg is False:
        return False

    if arg is True:
        return True

    arg = arg.lower().strip()
    if arg in {'y', 'yes', 't', 'true', 'on', '1'}:
        return True
    elif arg in {'n', 'no', 'f', 'false', 'off', '0'}:
        return False
    else:
        raise ValueError("invalid boolean value {!r:}".format(arg))


##########################################################################
## Environment
##########################################################################

# Names
NEVIS = "nevis.cs.umd.edu"
HYPERION = "hyperion.cs.umd.edu"
LAGOON = "lagoon.cs.umd.edu"

# Paths
workspace = "/data/honu"

# Fabric Env
env.colorize_errors = True
env.hosts = [NEVIS, HYPERION, LAGOON]
env.roledefs = {
    "client": {HYPERION, LAGOON},
    "server": {NEVIS},
}
env.user = "benjamin"
env.client_keys = KeyGen()


def multiexecute(task, n, host, *args, **kwargs):
    """
    Execute the task n times on the specified host. If the task is parallel
    then this will be parallel as well. All other args are passed to execute.
    """
    # Do nothing if n is zero or less
    if n < 1: return

    # Return one execution of the task with the given host
    if n == 1:
        return execute(task, host=host, *args, **kwargs)

    # Otherwise create a lists of hosts, don't dedupe them, and execute
    hosts = [host]*n
    with settings(dedupe_hosts=False):
        execute(task, hosts=hosts, *args, **kwargs)


##########################################################################
## Honu Commands
##########################################################################

def _serve(relax=False, uptime="45s"):
    relax = strpbool(relax)

    with cd(workspace):
        cmd = "honu serve -s -u {} -w server.jsonl".format(uptime)
        if relax:
            cmd += " -r"
        run(cmd)


@parallel
@roles('server')
def serve(relax=False, uptime="45s"):
    _serve(relax, uptime)


def _workload(multikey=False, duration="30s", server=NEVIS):
    multikey = strpbool(multikey)

    # Add the default port to the server
    if ":" not in server:
        server = server + ":3264"

    with cd(workspace):
        ckey = env.client_keys() if multikey else "FOO"
        cmd = "honu run -A -a {} -d {} -k {} -w client.jsonl".format(
                    server, duration, ckey
                )
        run(cmd)


@parallel
@roles('client')
def workload(multikey=False, duration="30s", server=NEVIS):
    _workload(multikey, duration, server)


@parallel
def experiment(relax=False,multikey=False,procs=2):
    procs = int(procs)
    cprocs = procs / 2
    if procs % 2 == 1 and env.host == LAGOON:
        cprocs += 1

    if env.host in env.roledefs['client']:
        multiexecute(_workload, cprocs, env.host, multikey=multikey)

    elif env.host in env.roledefs['server']:
        execute(_serve, host=env.host, relax=relax)


@parallel
def getmerge(localpath="."):
    local = os.path.join(localpath, "%(host)s", "%(path)s")
    remote = "client.jsonl" if env.host in env.roledefs['client'] else "server.jsonl"
    remote = os.path.join("/data/honu", remote)
    get(remote, local)
