#!/usr/bin/python
# -*- coding: utf-8 -*-

# 用法：
# 请在将如下的 json 配置放置在 /etc/mysql-mon.json
# 自己修改相应的参数，并且在这台机器上启用这个插件
# {
#     "host":     "localhost",
#     "port":     3306,
#     "database": "foo",
#     "user":     "foo",
#     "password": "bar",
# }

from __future__ import absolute_import

# -- prioritized --
import sys
import os.path
sys.path.append(os.path.join(os.path.dirname(__file__), '../libs'))

# -- stdlib --
import itertools
import json
import re
import socket
import time

# -- third party --
import pymysql


# -- own --
# -- code --
class Piper(object):
    def __init__(self, f):
        self.f = f

    def __ror__(self, arg):
        return self.f(arg)

    @classmethod
    def make(cls, f):
        return cls(f)


@Piper.make
def mapping(c):
    assert len(c.description) == 2
    return dict(c)


@Piper.make
def one(c):
    colnames = zip(*c.description)[0]
    row = c.fetchone()
    if not row:
        return None

    return dict(zip(itertools.cycle(colnames), row))


@Piper.make
def lower_keys(d):
    return {k.lower(): v for k, v in d.iteritems()}


endpoint = socket.gethostname()
ts = int(time.time())

{
    "host":     "localhost",
    "port":     3389,
    "database": "foo",
    "user":     "foo",
    "password": "bar",
}

try:
    kwargs = json.loads(open('/etc/mysql-mon.json').read())
except IOError:
    print '[]'
    sys.exit(0)

if kwargs['host'] not in ('localhost', '127.0.0.1'):
    endpoint = kwargs['host']

conn = pymysql.connect(**kwargs)
c = conn.cursor()


def Q(sql):
    c.execute(sql)
    return c


def norm(v):
    if isinstance(v, (int, float)):
        return v

    if isinstance(v, basestring):
        if v.lower() in ('off', 'false'):
            return 0
        elif v.lower() in ('on', 'true'):
            return 1
        elif v.isdigit():
            return float(v)
        else:
            return None

    return None

rst = []

status = Q('SHOW /*!50001 GLOBAL */ STATUS') | mapping | lower_keys
rst.extend([('mysql.status.' + k, norm(v)) for k, v in status.iteritems()])

slave = Q('SHOW SLAVE STATUS') | one
if slave:
    slave |= lower_keys
    rst.extend([('mysql.slave.' + k, norm(v)) for k, v in slave.iteritems()])

innodb = Q('SHOW /*!50000 ENGINE */ INNODB STATUS') | one | lower_keys
innodb = re.match(r'.+Mutex spin waits (?P<mutex_spin_waits>\d+), rounds (?P<mutex_spin_rounds>\d+), OS waits (?P<mutex_os_waits>\d+).+RW-shared spins (?P<rw_shared_spins>\d+), rounds (?P<rw_shared_rounds>\d+), OS waits (?P<rw_shared_os_waits>\d+).+RW-excl spins (?P<rw_excl_spins>\d+), rounds (?P<rw_excl_rounds>\d+), OS waits (?P<rw_os_waits>\d+)', innodb['status'], re.DOTALL).groupdict()

rst.extend([('mysql.innodb.' + k, norm(v)) for k, v in innodb.iteritems()])
c.close()

print json.dumps([{
    'metric': k,
    'endpoint': endpoint,
    'timestamp': ts,
    'step': 30,
    'value': v,
} for k, v in rst if v is not None])
