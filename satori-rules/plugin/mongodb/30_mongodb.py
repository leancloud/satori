#!/usr/bin/python
# -*- coding: utf-8 -*-
# -- prioritized --
import sys
import os.path
sys.path.append(os.path.join(os.path.dirname(__file__), '../libs'))

# -- stdlib --
import socket
import json
import time
import pymongo

# -- third party --
# -- own --

# -- code --
endpoint = socket.gethostname()
ts = int(time.time())

cli = pymongo.MongoClient('localhost', 27018, connectTimeoutMS=1000)
db = cli.get_database('test')

try:
    status = db.command('serverStatus', locks=0, recordStats=0)
except Exception:
    print json.dumps({
        'metric': 'mongodb.collect_success',
        'endpoint': endpoint,
        'timestamp': ts,
        'step': 30,
        'value': 0,
    })
    sys.exit(1)


metrics = [
    u'asserts.msg:COUNTER',
    u'asserts.regular:COUNTER',
    u'asserts.rollovers:COUNTER',
    u'asserts.user:COUNTER',
    u'asserts.warning:COUNTER',
    u'backgroundFlushing.average_ms:GAUGE',
    u'backgroundFlushing.flushes:COUNTER',
    u'backgroundFlushing.last_ms:GAUGE',
    u'backgroundFlushing.total_ms:COUNTER',
    u'connections.available:GAUGE',
    u'connections.current:GAUGE',
    u'globalLock.activeClients.readers:GAUGE',
    u'globalLock.activeClients.total:GAUGE',
    u'globalLock.activeClients.writers:GAUGE',
    u'globalLock.currentQueue.readers:GAUGE',
    u'globalLock.currentQueue.total:GAUGE',
    u'globalLock.currentQueue.writers:GAUGE',
    u'globalLock.lockTime:COUNTER',
    u'globalLock.totalTime:COUNTER',
    u'mem.mapped:GAUGE',
    u'mem.mappedWithJournal:GAUGE',
    u'mem.resident:GAUGE',
    u'mem.virtual:GAUGE',
    u'metrics.cursor.open.noTimeout:GAUGE',
    u'metrics.cursor.open.pinned:GAUGE',
    u'metrics.cursor.open.tota:GAUGE',
    u'metrics.cursor.timedOut:COUNTER',
    u'metrics.document.deleted:COUNTER',
    u'metrics.document.inserted:COUNTER',
    u'metrics.document.returned:COUNTER',
    u'metrics.document.updated:COUNTER',
    u'metrics.queryExecutor.scanned:COUNTER',
    u'metrics.queryExecutor.scannedObjects:COUNTER',
    u'metrics.record.moves:COUNTER',
    u'network.bytesIn:COUNTER',
    u'network.bytesOut:COUNTER',
    u'network.numRequests:COUNTER',
    u'opcounters.command:COUNTER',
    u'opcounters.delete:COUNTER',
    u'opcounters.getmore:COUNTER',
    u'opcounters.insert:COUNTER',
    u'opcounters.query:COUNTER',
    u'opcounters.update:COUNTER',
    u'opcountersRepl.command:COUNTER',
    u'opcountersRepl.delete:COUNTER',
    u'opcountersRepl.getmore:COUNTER',
    u'opcountersRepl.insert:COUNTER',
    u'opcountersRepl.query:COUNTER',
    u'opcountersRepl.update:COUNTER',
    u'repl.ismaster:GAUGE',
    u'collect_success:GAUGE',
]


def deepget(d, fields):
    fields = fields.split('.')
    v = d
    for f in fields:
        v = v.get(f)
        if v is None:
            return v

    return v

identity = lambda x: x
metric_conv = {
    'repl.ismaster': int,
    'collect_success': lambda x: 1,
}

conv = lambda m, v: metric_conv.get(m, identity)(v)

rst = []

for m in metrics:
    k, t = m.split(':')
    v = deepget(status, k)
    if v is None:
        continue

    rst.append({
        'metric': 'mongodb.%s' % k,
        'endpoint': endpoint,
        'timestamp': ts,
        'step': 30,
        'value': conv(k, v),
    })

print json.dumps(rst)
