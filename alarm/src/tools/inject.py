# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- stdlib --
import argparse
import json
import time

# -- third party --
import redis

# -- own --


# -- code --
def main():
    parser = argparse.ArgumentParser('inject')
    parser.add_argument('--redis',    type=str, default='redis://localhost:6379/0')
    parser.add_argument('--id',       type=str, default='injected-test-event-id')
    parser.add_argument('--endpoint', type=str, default='test-endpoint')
    parser.add_argument('--note',     type=str, default="SpaceX have just blasted Facebook's satellite!")
    parser.add_argument('--status',   type=str.upper, default='PROBLEM')
    parser.add_argument('--level',    type=int, default=3)
    parser.add_argument('--metric',   type=str, default='facebook.satellite.healthy')
    parser.add_argument('--time',     type=int, default=int(time.time()))
    parser.add_argument('--tags',     type=eval, default={})
    parser.add_argument('--actual',   type=float, default=0)
    parser.add_argument('--expected', type=float, default=1)
    parser.add_argument('--teams',    type=str, nargs='+', default=['operation'])

    options = parser.parse_args()

    ev = {
        u'status':   options.status,
        u'tags':     options.tags,
        u'metric':   options.metric,
        u'groups':   options.teams,
        u'id':       options.id,
        u'endpoint': options.endpoint,
        u'actual':   options.actual,
        u'level':    options.level,
        u'note':     options.note,
        u'time':     options.time,
        u'expected': options.expected,
    }

    r = redis.from_url(options.redis)
    r.rpush("satori-events:%s" % options.level, json.dumps(ev))
