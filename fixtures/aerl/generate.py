#!/usr/bin/env python

import os
import json
import argparse

from string import ascii_uppercase
from collections import defaultdict

BASE   = os.path.dirname(__file__)
DATA   = os.path.join(BASE, "data")
CONFIG = os.path.join(BASE, "config.json")
HOSTS  = os.path.join(BASE, "hosts.json")


def load_hosts_in_locations(path=HOSTS, regions=None):
    hosts = defaultdict(list)
    with open(path, 'r') as f:
        for hostname in json.load(f).keys():
            loc = " ".join(hostname.split("-")[1:-1])
            hosts[loc].append(hostname)

    if regions is not None:
        return {
            region: hosts[region]
            for region in regions
        }

    return hosts


def main(args):
    # Load hosts
    hosts = load_hosts_in_locations(path=args.hosts, regions=args.regions)

    # Load config
    with open(args.config, 'r') as f:
        config = json.load(f)

    # Create hosts for replicas and clients
    replicas = []
    clients = []

    for loc, names in hosts.items():
        if args.replicas > len(names):
            raise ValueError(
                "not enough hosts in {} for {} replicas".format(loc, args.replicas)
            )

        if args.clients > len(names):
            raise ValueError(
                "not enough hosts in {} for {} clients".format(loc, args.cliens)
            )

        replicas += names[:args.replicas]
        clients += names[:args.clients]

    config['replicas']['hosts'] = replicas
    config['clients']['hosts'] = {}

    for idx, client in enumerate(clients):
        config['clients']['hosts'][client] = ascii_uppercase[idx%26]

    if args.uniform or args.annealing:
        config['replicas'].pop('epsilon', None)

    if args.uniform:
        config['replicas']['config']['bandit'] = 'uniform'
    elif args.annealing:
        config['replicas']['config']['bandit'] = 'annealing'
    elif args.epsilon:
        config['replicas']['config']['bandit'] = 'epsilon'
        config['replicas']['config']['epsilon'] = args.epsilon

    if args.outpath is not None:
        with open(args.outpath, 'w') as o:
            json.dump(config, o, indent=2)

    else:
        print(json.dumps(config, indent=2))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description="generates experimental configurations"
    )

    parser.add_argument(
        '-C', '--config', default=CONFIG, metavar='PATH',
        help='location of base configuration file',
    )

    parser.add_argument(
        '-H', '--hosts', default=HOSTS, metavar='PATH',
        help='location of hosts.json file'
    )

    parser.add_argument(
        '-R', '--regions', default=None, nargs='*',
        help='specify regions to use, default is all available',
    )

    parser.add_argument(
        '-o', '--outpath', default=None, metavar='PATH',
        help='path to write configuration out to',
    )

    parser.add_argument(
        '-r', '--replicas', default=3, metavar='N', type=int,
        help='number of replicas per region'
    )

    parser.add_argument(
        '-c', '--clients', default=1, metavar='N', type=int,
        help='number of clients per region'
    )

    bandit = parser.add_mutually_exclusive_group(required=True)

    bandit.add_argument(
        '-e', '--epsilon', type=float, metavar='E',
        help='use epsilon greedy bandit with specified epsilon',
    )

    bandit.add_argument(
        '-A', '--annealing', action='store_true',
        help="create an annealing bandit experiment",
    )

    bandit.add_argument(
        '-U', '--uniform', action='store_true',
        help="create an uniform bandit experiment",
    )

    args = parser.parse_args()
    # try:
    main(args)
    # except Exception as e:
    #     parser.error(str(e))
