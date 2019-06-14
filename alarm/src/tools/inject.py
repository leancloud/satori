# -*- coding: utf-8 -*-


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
        'status':   options.status,
        'tags':     options.tags,
        'metric':   options.metric,
        'groups':   options.teams,
        'id':       options.id,
        'endpoint': options.endpoint,
        'actual':   options.actual,
        'level':    options.level,
        'note':     options.note,
        'time':     options.time,
        'expected': options.expected,
    }

    r = redis.from_url(options.redis)
    r.rpush("satori-events:%s" % options.level, json.dumps(ev))
