#!/usr/bin/python
# -*- coding: utf-8 -*-

# -- stdlib --
import json
import socket
import subprocess
import time

# -- third party --
# -- own --

# -- code --
endpoint = socket.gethostname()
ts = int(time.time())

cmdline = ' | '.join([
    'pgrep -f QuorumPeerMain',
    'xargs',
    'sed "s/ /|/g"',
    'xargs -I{} bash -c \'(sudo netstat -tulnp | grep -E "{}")\'',
    "awk '{print $4}'",
    "awk -F: '{print $NF}'",
])

proc = subprocess.Popen(['/bin/bash', '-c', cmdline], stdout=subprocess.PIPE)
ports = map(int, proc.stdout.read().strip().split())
ports = [p for p in ports if p < 3000]

rst = []

metrics = [
    "collect_success::lambda v: 1::GAUGE",
    "avg_latency::int::GAUGE",
    "max_latency::int::GAUGE",
    "min_latency::int::GAUGE",
    "packets_received::int::COUNTER",
    "packets_sent::int::COUNTER",
    "num_alive_connections::int::GAUGE",
    "outstanding_requests::int::GAUGE",
    "server_state::lambda v:['follower', 'leader'].index(v)::GAUGE",
    "znode_count::int::GAUGE",
    "watch_count::int::GAUGE",
    "ephemerals_count::int::GAUGE",
    "approximate_data_size::int::GAUGE",
    "open_file_descriptor_count::int::GAUGE",
    "max_file_descriptor_count::int::GAUGE",
]

for port in ports:
    try:
        s = socket.socket()
        s.connect(('0.0.0.0', port))
        s.sendall('mntr\r\n')
        data = []
        while True:
            d = s.recv(2048)
            if not d:
                break
            data.append(d)
        s.close()
    except:
        continue

    raw = ''.join(data).strip().split('\n')
    raw = {k: v for k, v in [i.split()[:2] for i in raw]}

    for i in metrics:
        m, value_type, counter_type = i.split('::')
        v = eval(value_type)(raw.get('zk_' + m))
        rst.append({
            'metric': 'zookeeper.%s' % m,
            'endpoint': endpoint,
            'timestamp': ts,
            'step': 60,
            'value': v,
            'tags': {'port': str(port)},
        })

print json.dumps(rst)
