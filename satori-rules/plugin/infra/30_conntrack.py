#!/usr/bin/python
# -*- coding: utf-8 -*-

# -- stdlib --
import json
import os
import sys
import time

# -- third party --
# -- own --

# -- code --
COUNT = '/proc/sys/net/netfilter/nf_conntrack_count'
MAX = '/proc/sys/net/netfilter/nf_conntrack_max'

if not os.path.exists(COUNT):
    print '[]'
    sys.exit(0)

ts = int(time.time())

count = int(open(COUNT).read())
max = int(open(MAX).read())

metric = [
    {
        "metric": "net.netfilter.conntrack.used",
        "timestamp": ts,
        "value": count,
    }, {
        "metric": "net.netfilter.conntrack.max",
        "timestamp": ts,
        "value": max,
    }, {
        "metric": "net.netfilter.conntrack.used_ratio",
        "timestamp": ts,
        "value": 1.0 * count / max,
    }
]

print json.dumps(metric)
