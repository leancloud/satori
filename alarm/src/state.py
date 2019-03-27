# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- stdlib --
# -- third party --
# -- own --
from utils import instantiate


# -- code --
@instantiate
class State(object):
    __slots__ = [
        'config',
        'options',
        'teams',
        'users',
        'strategies',
        'alarms',
        'backends',
        'userconf',
    ]
