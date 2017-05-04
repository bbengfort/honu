#!/usr/bin/env python3

import os
import re
import glob

import seaborn as sns
import matplotlib.pyplot as plt

sns.set_context('poster')
sns.set_style('whitegrid')
sns.set_palette('Blues', 13)


def parse_file(path):
    pat = re.compile(r'^MEM\s+([\d\.]+)\s+([\d\.]+)$', re.I)
    start = None

    with open(path, 'r') as f:
        for line in f:
            match = pat.match(line.strip())
            if match:
                row = tuple(map(float, match.groups()))
                if start is None:
                    start = row[1]

                ts = row[1] - start
                yield ts, row[0]


def parse_series(paths):
    pat = re.compile(r'^mprofile-(\d+)-nodes.dat$', re.I)
    for path in paths:
        match = pat.match(os.path.basename(path))
        if match:
            name = "{} clients".format(int(match.group(1)))
            series = list(parse_file(path))
            yield name, series


def mprof_plot(paths):
    for name, series in parse_series(paths):
        x = [row[0] for row in series]
        y = [row[1] for row in series]
        plt.plot(x, y, label=name)

    plt.title("Server Memory Usage with for Concurrent Clients")
    plt.ylabel("memory (MiB)")
    plt.xlabel("seconds")
    plt.legend(loc='best')



if __name__ == '__main__':
    mprof_plot(glob.glob('mprofile-*'))
    plt.show()
