# -*- coding: utf-8 -*-
from __future__ import absolute_import

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
