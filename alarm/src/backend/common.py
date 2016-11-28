# -*- coding: utf-8 -*-

# -- stdlib --
import datetime
import shlex

# -- third party --
# -- own --

# -- code --
BACKENDS = {}


def register_backend(f):
    BACKENDS[f.__name__] = f
    return f


def from_string(s):
    return BACKENDS[s]
