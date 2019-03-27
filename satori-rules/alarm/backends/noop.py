# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
# -- own --
from backend import Backend


# -- code --
class NoopBackend(Backend):
    pass


EXPORT = NoopBackend
