#!/usr/bin/python
# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- stdlib --
import glob
import json
import socket
import subprocess
import time

# -- third party --
import requests

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

nodes = '\n'.join([i['value'] for i in resp.json()])

p = subprocess.Popen(['/usr/bin/fping'], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=open('/dev/null', 'w'))
p.stdin.write(nodes)
p.stdin.write('\n')
p.stdin.close()
result = p.stdout.read()
result = [i.split(' is ') for i in result.strip().split('\n')]

metric = [{
    "metric": "ping.alive",
    "endpoint": k,
    "timestamp": ts,
    "step": 30,
    "value": ("unreachable", "alive").index(v),
    "tags": {"from": endpoint},
} for k, v in result]

print json.dumps(metric)
