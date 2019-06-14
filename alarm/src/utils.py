# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
import gevent

# -- own --


# -- code --
def instantiate(cls):
    return cls()


def spawn_autorestart(*args, **kwargs):
    def restart(g):
        gevent.sleep(1)
        spawn_autorestart(*args, **kwargs)

    gevent.spawn(*args, **kwargs).link(restart)


def status2emoji(s):
    return {
        'PROBLEM': '😱',
        'EVENT': '😱',
        'OK': '😅',
    }.get(s, s)
