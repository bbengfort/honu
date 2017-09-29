#!/usr/bin/env python3
# mkenv
# Creates the environment helper scripts in the envs directory.
#
# Author:  Benjamin Bengfort <benjamin@bengfort.com>
# Created: Tue Sep 19 11:01:35 2017 -0400
#
# ID: mkenv.py [] benjamin@bengfort.com $

"""
Creates the environment helper scripts in the envs directory.
"""

##########################################################################
## Imports
##########################################################################

import os
import json
import argparse

# Paths on Disk
BASE  = os.path.dirname(__file__)
ENVS  = os.path.join(BASE, "envs")
HONU  = os.path.abspath(os.path.join(ENVS, "honu.sh"))
NAMES = [
    "alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel",
    "indigo", "juliet", "kilo", "lima", "mike", "november", "oscar", "papa",
    "quebec", "romeo", "sierra", "tango", "unicorn", "victor", "whiskey",
    "xray", "yankee", "zulu",
]

# Configuration Defaults
PORT  = 3264
RELAX = False
DELAY = "1s"
STANDALONE = False
UPTIME  = "40s"
RESULTS = os.path.abspath(os.path.join(BASE, "metrics.json"))
BANDIT  = "uniform"
EPSILON = 0.1


def write_env(name, **kwargs):

    def str_env(key,val):
        if val:
            return "export {}=\"{}\"\n".format(key.upper(), val)

    def list_env(key,val):
        if val:
            items = ",".join(map(str, val))
            return "export {}=\"{}\"\n".format(key.upper(), items)

    def bool_env(key, val):
        if val:
            return "export {}=true\n".format(key.upper())
        return "export {}=false\n".format(key.upper())

    def int_env(key, val):
        return "export {}={}\n".format(key.upper(), val)

    def float_env(key, val):
        return "export {}={:0.2f}\n".format(key.upper(), val)


    # Writers for various keys
    writers = {
        "HONU_SERVER_ADDR": str_env,
        "HONU_SEQUENTIAL_CONSISTENCY": bool_env,
        "HONU_PROCESS_ID": int_env,
        "HONU_PEERS": list_env,
        "HONU_ANTI_ENTROPY_DELAY": str_env,
        "HONU_STANDALONE_MODE": bool_env,
        "HONU_SERVER_UPTIME": str_env,
        "HONU_SERVER_RESULTS": str_env,
        "HONU_BANDIT_STRATEGY": str_env,
        "HONU_BANDIT_EPSILON": float_env,
    }

    path = os.path.join(ENVS, name+".sh")
    with open(path, 'w') as f:
        # Write strings to the env file
        for key, val in kwargs.items():
            line = writers[key](key, val)
            f.write(line)

        # Source the HONU shell script last
        f.write("\nsource {}\n".format(HONU))


def main(args):
    if args.n[0] < 1 or args.n[0] > len(NAMES):
        raise ValueError("specify between 1 and {} hosts".format(len(NAMES)))

    kwargs = {
        "HONU_SERVER_ADDR": "",
        "HONU_SEQUENTIAL_CONSISTENCY": args.relax,
        "HONU_PROCESS_ID": 0,
        "HONU_PEERS": [],
        "HONU_ANTI_ENTROPY_DELAY": args.delay,
        "HONU_STANDALONE_MODE": args.standalone,
        "HONU_SERVER_UPTIME": args.uptime,
        "HONU_SERVER_RESULTS": args.results,
        "HONU_BANDIT_STRATEGY": args.bandit,
        "HONU_BANDIT_EPSILON": args.epsilon,
    }

    for idx in range(args.n[0]):
        name   = NAMES[idx]
        kwargs["HONU_PROCESS_ID"] = idx + 1
        kwargs["HONU_SERVER_ADDR"] = "{}:{}".format(args.addr, PORT + idx)
        kwargs["HONU_PEERS"] = [
            ":{}".format(jdx+PORT)
            for jdx in range(args.n[0])
            if idx != jdx
        ]
        write_env(name, **kwargs)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description="create the environment files for each replica",
        epilog="part of the Honu experimental toolkit"
    )

    parser.add_argument(
        "n", nargs=1, type=int, help="number of replicas to create",
    )
    parser.add_argument(
        "-a", "--addr", default="", help="ipaddress of the server",
    )
    parser.add_argument(
        "-r", "--relax", action="store_true", default=RELAX,
        help="relax to sequential consistency",
    )
    parser.add_argument(
        "-d", "--delay", default=DELAY,
        help="parseable duration of anti-entropy delay",
    )
    parser.add_argument(
        "-s", "--standalone", default=STANDALONE, action="store_true",
        help="disable replication and run in standalone mode",
    )
    parser.add_argument(
        "-u", "--uptime", default=UPTIME,
        help="parseable duration to shut the server down after",
    )
    parser.add_argument(
        "-w", "--results", default=RESULTS,
        help="path on disk to write JSON stats to on shutdown",
    )
    parser.add_argument(
        "-b", "--bandit", default=BANDIT,
        help="bandit strategy for random peer selection",
    )
    parser.add_argument(
        "-e", "--epsilon", default=EPSILON, type=float,
        help="value of epsilon for epsilon greedy selection",
    )

    args = parser.parse_args()
    main(args)
