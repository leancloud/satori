#!/usr/bin/python
# -*- coding: utf-8 -*-

# -- stdlib --
from collections import Counter
import json
import sys
import time
import urllib2

# -- third party --
# -- own --

# -- code --
ts = int(time.time())

result = [{
    "metric": "nginx.collect_success",
    "timestamp": ts,
    "value": 1,
}]

try:
    info = json.loads(urllib2.urlopen('http://127.0.0.1/status?format=json').read())
    info = info['servers']['server']
    total = Counter([i['upstream'] for i in info])
    healthy = Counter([i['upstream'] for i in info if i['status'] == 'up'])
    upstreams = total.keys()
except Exception:
    result[0]['value'] = 0
    print json.dumps(result)
    sys.exit(1)


for u in total.keys():
    result.extend([
        {
            "metric": "nginx.upstream.total",
            "timestamp": ts,
            "value": total[u],
            "tags": {"upstream": u},
        },
        {
            "metric": "nginx.upstream.healthy",
            "timestamp": ts,
            "value": healthy[u],
            "tags": {"upstream": u},
        },
        {
            "metric": "nginx.upstream.healthy.ratio",
            "timestamp": ts,
            "value": 1.0 * healthy[u] / total[u],
            "tags": {"upstream": u},
        },
    ])

print json.dumps(result)
