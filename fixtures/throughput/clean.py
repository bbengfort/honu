#!/usr/bin/env python

import os
import re
import csv
import glob


def rename_hosts(path):
    """
    Rename the long form of the hosts to a more manageable form.
    """
    # Regex pattern to match convention
    pat = re.compile(r'^bengfort@bbc(\d+)\.cs\.umd\.edu$', re.I)

    for name in os.listdir(path):
        match = pat.match(name)
        if match:
            num = match.groups()[0]
            newname = "bbc{:0>2}".format(num)
            os.rename(os.path.join(path, name), os.path.join(path, newname))


def rename_mprofile(path, ascending=False):
    """
    Rename the mprofile .dat files to the number of nodes. If ascending is
    True, then will count up number of nodes, otherwise will count down.
    """
    rng = range(1, 26, 2) if ascending else range(25, 0, -2)

    for idx, name in zip(rng, os.listdir(path)):
        newname = os.path.join(path, "mprofile-{:0>2}-nodes.dat".format(idx))
        os.rename(os.path.join(path, name), newname)


def merge_results(paths, outpath='results.csv', fields=None):
    """
    Merge multiple CSV results files from different nodes into a single csv.
    If fields is specified then only compile a subset of the dimensions.
    """
    # Regex pattern to match filenames
    pat = re.compile(r'^results-(\d+)-nodes.csv$', re.I)

    # Fields or all the fields
    fields = fields or ("clients", "msg", "key", "version", "timestamp", "latency (ns)", "bytes", "success")

    # Open an writer to the output file
    with open(outpath, 'w') as out:
        writer = csv.DictWriter(out, fieldnames=fields)
        writer.writeheader()

        # Loop through the paths and find the results files
        for path in paths:
            for name in os.listdir(path):
                match = pat.match(name)
                if not match: continue

                nodes = int(match.groups()[0])
                with open(os.path.join(path, name), 'r') as f:
                    reader = csv.DictReader(f)
                    for row in reader:
                        row["clients"] = nodes
                        writer.writerow(row)


def throughputs(paths, outpath='throughput.csv'):
    # Regex pattern to match filenames
    pat = re.compile(r'^results-(\d+)-nodes.csv$', re.I)

    # Open an writer to the output file
    with open(outpath, 'w') as out:
        writer = csv.DictWriter(out, fieldnames=('clients', "messages", "latency (ns)", "bytes"))
        writer.writeheader()

        # Loop through the paths and find the results files
        for path in paths:
            for name in os.listdir(path):
                match = pat.match(name)
                if not match: continue

                nodes = int(match.groups()[0])
                msgs  = 0
                latency = 0
                nbytes = 0
                with open(os.path.join(path, name), 'r') as f:
                    reader = csv.DictReader(f)
                    for row in reader:
                        try:
                            msgs += 1
                            latency += int(row['latency (ns)'])
                            nbytes += int(row['bytes'])
                        except:
                            continue

                row = {
                    "clients": nodes,
                    "messages": msgs,
                    "latency (ns)": latency,
                    "bytes": nbytes,
                }

                writer.writerow(row)


if __name__ == '__main__':
    throughputs(glob.glob("bbc*"))
