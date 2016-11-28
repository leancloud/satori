#!/usr/bin/python
# -*- coding: utf-8 -*-

# -- stdlib --
import json
import socket
import re
import telnetlib
import time

# -- third party --
# -- own --

# -- code --
QUEUE_METRIC = re.compile(r'^queue_(.*?)_(items|bytes|total_items|logsize|expired_items|mem_items|mem_bytes|age|discarded|waiters|open_transactions|transactions|canceled_transactions|total_flushes|journal_rewrites|journal_rotations|creates|deletes|expires)$')

endpoint = socket.gethostname()
ts = int(time.time())

ports = [22133]

rst = []

for port in ports:
    try:
        conn = telnetlib.Telnet('0.0.0.0', port)
        conn.write('stats\r\nquit\r\n')
        lines = conn.read_until('END')
    except:
        continue

    lines = lines.split('\r\n')
    assert lines[-1].strip() == 'END'
    lines = lines[:-1]
    conn.close()

    for l in lines:
        _, k, v = l.strip().split()
        if k in ('pid', 'uptime', 'version', 'libevent', 'time'):
            continue

        v = float(v)
        m = QUEUE_METRIC.match(k)
        if m:
            queue, metric = m.groups()
            m2 = re.match(r'(^.*?)-(\d+)$', queue)
            if m2:
                queue, num = m2.groups()
            else:
                num = '-'
            metric = 'kestrel_queue.%s' % metric.replace('_', '.')
            t = {'port': str(port), 'queue': queue, 'num': num}
        else:
            metric = 'kestrel.%s' % k
            t = {'port': str(port)}

        rst.append({
            'metric': metric,
            'endpoint': endpoint,
            'timestamp': ts,
            'step': 60,
            'value': v,
            'tags': t,
        })

    del k, v

print json.dumps(rst)
