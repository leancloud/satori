#!/usr/bin/python
# -*- coding: utf-8 -*-
from __future__ import absolute_import


# -- stdlib --
import json
import re
import subprocess
import time

# -- third party --

# -- own --

# -- code --
ts = int(time.time())

proc = subprocess.Popen(['/usr/sbin/unbound-control', 'stats'], stdout=subprocess.PIPE)
stats = {
    match[0]: float(match[1])
    for match in re.findall(r'(.*)\=(.*)', proc.stdout.read(), re.MULTILINE)
}

rst = {
    'uptime': stats['time.up'],
    'queries.total': stats['total.num.queries'],
    'queries.pending': stats['total.requestlist.current.all'],
    'queries.hit_rate': (stats['total.num.cachehits'] / stats['total.num.queries']) * 100,
}

print json.dumps([
    {
        "metric": "unbound.{}".format(k),
        "timestamp": ts,
        "value": int(v),
    }
    for k, v in rst.items()
])
