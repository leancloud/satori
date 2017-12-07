#!/usr/bin/python
# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- prioritized --
import sys
import os.path
sys.path.append(os.path.join(os.path.dirname(__file__), '../libs'))
from gevent import monkey
monkey.patch_all()

# -- stdlib --
import glob
import json
import socket
import time

# -- third party --
from gevent.pool import Pool
import requests
import gevent

# -- own --

# -- code --
endpoint = socket.gethostname()
ts = int(time.time())
key = glob.glob('/var/lib/puppet/ssl/private_keys/%s*' % endpoint)[:1]
cert = glob.glob('/var/lib/puppet/ssl/certs/%s*' % endpoint)[:1]

resp = requests.get('https://puppet:9081/v3/facts',
    params={'query': json.dumps(['=', 'name', 'hostname'])},
    cert=cert+key,
    verify=False,
)

nodes = [i['value'] for i in resp.json()]


def detect(node):
    for i in xrange(3):
        try:
            resp = requests.get('http://%s:1988/v1/ping' % node, timeout=5).json()
            return (node, resp['result'] == 'pong')
        except Exception:
            gevent.sleep(1)
            continue

    return (node, False)


metric = [{
    "metric": "agent.ping",
    "endpoint": node,
    "timestamp": ts,
    "step": 60,
    "value": int(v),
    "tags": {"from": endpoint},
} for node, v, in Pool(60).imap_unordered(detect, nodes)]

print json.dumps(metric)
