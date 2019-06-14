#!/usr/bin/python
# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- stdlib --
import ctypes
import json
import time

# -- third party --
# -- own --

# -- code --
ts = int(time.time())

libc = ctypes.CDLL('libc.so.6')
sz = libc.klogctl(10, 0, 0)  # man 2 syslog
buf = ctypes.create_string_buffer(sz+1)
libc.klogctl(3, buf, sz)
msgs = buf.value

rst = [{
    'metric': 'kernel.dmesg.bug',
    'timestamp': ts,
    'step': 60,
    'value': msgs.count('BUG:'),
}, {
    'metric': 'kernel.dmesg.io_error',
    'timestamp': ts,
    'step': 60,
    'value': msgs.count('I/O error'),
}]

print json.dumps(rst)
