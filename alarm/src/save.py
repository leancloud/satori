# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- stdlib --
import logging

# -- third party --
import gevent
import redis

# -- own --
from state import State
from utils import spawn_autorestart

# -- code --
log = logging.getLogger('save')


def dump():
    r = redis.from_url(State.config['redis'])
    r.set('satori-saved-alarms', State.alarms.dumps())


def load():
    r = redis.from_url(State.config['redis'])
    s = r.get('satori-saved-alarms')
    if not r:
        return

    try:
        s = State.alarms.loads(s)
    except Exception:
        return


def start_periodically_dump():
    @spawn_autorestart
    def do_periodically_dump():
        while True:
            gevent.sleep(5)
            try:
                dump()
            except Exception:
                log.exception('Error dumping satori events')
