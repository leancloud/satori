#!/usr/bin/python
# -*- coding: utf-8 -*-

# -- stdlib --
import json
import os

# -- third party --
# -- own --

# -- code --
rst = os.system('swapon -s | grep file >/dev/null 2>/dev/null')

print json.dumps([{
    "metric": "mem.swaponfile",
    "step": 600,
    "value": float(rst == 0),
}])
