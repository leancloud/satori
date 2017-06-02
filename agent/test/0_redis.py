#!/usr/bin/python -u
# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- prioritized --

# -- stdlib --
import socket
import time
import json
# -- third party --

# -- own --

# -- code --
endpoint = socket.gethostname()
ts = int(time.time())
rst = []

for k in range(100000):
    time.sleep(1)
    print json.dumps([{
        'metric': 'redis.%s' % k,
        'endpoint': endpoint,
        'timestamp': ts,
        'step': 30,
        'value': k,
    }])
