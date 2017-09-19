#!/usr/bin/python
# -*- coding: utf-8 -*-

# -- stdlib --
import json
import multiprocessing
import time

# -- third party --
# -- own --

# -- code --
nproc = multiprocessing.cpu_count()
_1min, _5min, _15min = map(lambda v: float(v) / nproc, open('/proc/loadavg').read().split()[:3])

ts = int(time.time())

metric = [
    {
        "metric": "load.1min.normalized",
        "timestamp": ts,
        "value": _1min,
    }, {
        "metric": "load.5min.normalized",
        "timestamp": ts,
        "value": _5min,
    }, {
        "metric": "load.15min.normalized",
        "timestamp": ts,
        "value": _15min,
    }
]
print json.dumps(metric)
