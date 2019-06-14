#!/usr/bin/python
# -*- coding: utf-8 -*-

# -- stdlib --
import json
import time

# -- third party --
# -- own --

# -- code --
ts = int(time.time())


lines = open('/proc/softirqs').read().strip().split('\n')
metrics = [i.split() for i in lines[1:]]
metrics = [{
    'metric': 'softirq.%s' % m[0][:-1].lower(),
    'value': float(v),
    'step': 30,
    'tags': {'cpu': str(i)},
} for m in metrics for i, v in enumerate(m[1:])]

print(json.dumps(metrics))
