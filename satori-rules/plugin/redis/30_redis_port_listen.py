#!/usr/bin/python
# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- prioritized --
import sys
import os.path
sys.path.append(os.path.join(os.path.dirname(__file__), '../libs'))

# -- stdlib --
import json
import subprocess

# -- third party --

# -- own --

# -- code --
NET_PORT_LISTEN = os.path.join(os.path.abspath(os.path.dirname(__file__)), '../_metric/net.port.listen')

proc = subprocess.Popen(['/bin/bash', '-c', "cat /etc/redis-*.conf | grep '^port ' | grep -Po '\d+'"], stdout=subprocess.PIPE)
ports = map(int, proc.stdout.read().split())

if not ports:
    ports = [6379]

reqs = [{
    '_metric': 'net.port.listen',
    '_step': 30,
    'port': i,
    'name': 'redis-port',
} for i in ports]

proc = subprocess.Popen([NET_PORT_LISTEN], stdin=subprocess.PIPE)
proc.stdin.write(json.dumps(reqs))
proc.stdin.close()
proc.wait()
