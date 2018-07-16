#!/usr/bin/python
# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- prioritized --
import sys
import os.path
sys.path.append(os.path.join(os.path.dirname(__file__), '../libs'))

# -- stdlib --
import json
import re
import socket
import subprocess
import time

# -- third party --
import redis

# -- own --

# -- code --
ts = int(time.time())

proc = subprocess.Popen(['/bin/bash', '-c', "ps axo cmd | grep 'redis-server '"], stdout=subprocess.PIPE)
ports = map(int, re.findall(r'redis-server .*?:(\d+) *$', proc.stdout.read(), re.MULTILINE))

interested = {
    'blocked_clients',
    'connected_clients',
    'connected_slaves',
    'evicted_keys',
    'expired_keys',
    'instantaneous_ops_per_sec',
    'keyspace_hits',
    'keyspace_misses',
    'loading',
    'mem_fragmentation_ratio',
    'pubsub_channels',
    'pubsub_patterns',
    'rdb_bgsave_in_progress',
    'rdb_last_bgsave_time_sec',
    'total_commands_processed',
    'total_connections_received',
    'used_cpu_sys',
    'used_cpu_sys_children',
    'used_cpu_user',
    'used_cpu_user_children',
    'used_memory',
    'used_memory_lua',
    'used_memory_rss',
    'maxmemory',
    'maxclients',
    'databases',
}


rst = []

for p in ports:
    r = redis.from_url('redis://0.0.0.0:%s' % p)
    try:
        config = r.config_get()
        info = r.info()
        info.update(config)
        rst.extend([{
            'metric': 'redis.%s' % k,
            'timestamp': ts,
            'step': 30,
            'value': float(info[k]),
            'tags': {'port': str(p)},
        } for k in interested if k in info])
    except redis.ConnectionError:
        pass

print json.dumps(rst)
