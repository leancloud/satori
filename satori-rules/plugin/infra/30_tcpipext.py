#!/usr/bin/env python
# -*- coding: utf-8 -*-
from __future__ import absolute_import, division, print_function, unicode_literals

# -- stdlib --
import json

# -- third party --
# -- own --

# -- code --
rst = []

with open('/proc/net/netstat') as stat:
    while True:
        header = stat.readline().strip()
        if not header:
            break

        values = stat.readline().strip()
        array = zip(header.split(), values.split())
        cat, array = array[0][0].strip(' :'), array[1:]

        for k, v in array:
            rst.append(('%s.%s' % (cat, k), int(v)))


rst = [{'metric': k, 'step': 30, 'value': v} for k, v in rst]
print(json.dumps(rst))
