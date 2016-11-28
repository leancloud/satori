#!/usr/bin/python
# -*- coding: utf-8 -*-

# -- stdlib --
import json
import socket
import subprocess
import telnetlib
import time

# -- third party --
# -- own --

# -- code --
endpoint = socket.gethostname()
ts = int(time.time())

proc = subprocess.Popen(['/bin/bash', '-c', r'''ps -ef |grep memcached|grep -v grep |sed -n 's/.* *-p *\([0-9]\{1,5\}\).*/\1/p' '''], stdout=subprocess.PIPE)
ports = map(int, proc.stdout.read().strip().split())

rst = []

for port in ports:
    try:
        conn = telnetlib.Telnet('0.0.0.0', port)
        conn.write('stats\r\nquit\r\n')
        lines = conn.read_until('END')
        lines = lines.split('\r\n')
        assert lines[-1] == 'END'
        conn.close()
    except:
        continue

    stats = dict([i.split()[1:] for i in lines[:-1]])
    [stats.pop(i, '') for i in ('pid', 'uptime', 'version', 'libevent', 'time')]
    stats = {k: float(v) for k, v in stats.items()}

    stats['usage'] = 100 * stats['bytes'] / stats['limit_maxbytes']

    def add_ratio(a, b):
        try:
            stats[a + '_ratio'] = 100 * stats[a] / (stats[a] + stats[b])
        except ZeroDivisionError:
            stats[a + '_ratio'] = 0

    add_ratio('get_hits', 'get_misses')
    add_ratio('incr_hits', 'incr_misses')
    add_ratio('decr_hits', 'decr_misses')
    add_ratio('delete_hits', 'delete_misses')
    add_ratio('cas_hits', 'cas_misses')
    add_ratio('touch_hits', 'touch_misses')

    rst.extend([{
        'metric': 'memcached.%s' % k,
        'endpoint': endpoint,
        'timestamp': ts,
        'step': 60,
        'value': v,
        'tags': {'port': str(port)},
    } for k, v in stats.items()])


print json.dumps(rst)
