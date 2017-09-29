#!/usr/bin/python
# -*- coding: utf-8 -*-

# -- stdlib --
import json
import socket
import time
import urllib2

# -- third party --
# -- own --

# -- code --
endpoint = socket.gethostname()
ts = int(time.time())

indices = urllib2.urlopen('http://127.0.0.1:9200/_cat/indices').read()

keys = 'health status index pri rep docs.count docs.deleted store.size pri.store.size'.split()
indices = [dict(zip(keys, i.split())) for i in indices.strip().split('\n')]

metrics = []

for i in indices:
    name = i['index']
    metrics.extend([
        ('%s.health' % name, ['red', 'yellow', 'green'].index(i['health'])),
        ('%s.pri' % name, int(i['pri'])),
        ('%s.rep' % name, int(i['rep'])),
        ('%s.docs.count' % name, int(i['docs.count'])),
        ('%s.docs.deleted' % name, int(i['docs.deleted'])),
    ])
    health = urllib2.urlopen('http://127.0.0.1:9200/_cluster/health/%s' % name).read()
    health = json.loads(health)
    health.pop('cluster_name', '')
    health.pop('status', '')
    health.pop('timed_out', '')
    metrics.extend([
        ('%s.%s' % (name, k), v)
        for k, v in health.items()
    ])

result = [{
    "metric": "es.%s" % k,
    "endpoint": endpoint,
    "timestamp": ts,
    "step": 60,
    "value": v,
} for k, v in metrics]

print json.dumps(result)
