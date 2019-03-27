#!/usr/bin/python
# -*- coding: utf-8 -*-

# -- prioritized --
import sys
import os.path
sys.path.append(os.path.join(os.path.dirname(__file__), '..'))

# -- stdlib --
from collections import Counter
import os
import json
import sys
import time

# -- third party --
import requests

# -- own --

# -- code --
ts = int(time.time())

result = [{
    "metric": "nginx.collect_success",
    "timestamp": ts,
    "step": 30,
    "value": 1,
}]

try:
    info = requests.get(
        'http://127.0.0.1:%s/status?format=json' % os.getenv('PORT_80', '80'),
        headers={'Host': 'health-check'}
    ).json()
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
            "step": 30,
            "value": total[u],
            "tags": {"upstream": u},
        },
        {
            "metric": "nginx.upstream.healthy",
            "timestamp": ts,
            "step": 30,
            "value": healthy[u],
            "tags": {"upstream": u},
        },
        {
            "metric": "nginx.upstream.healthy.ratio",
            "timestamp": ts,
            "step": 30,
            "value": 1.0 * healthy[u] / total[u],
            "tags": {"upstream": u},
        },
    ])

print json.dumps(result)
