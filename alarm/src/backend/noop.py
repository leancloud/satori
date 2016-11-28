# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
# -- own --
from backend.common import register_backend

# -- code --


@register_backend
def noop(conf, user, event):
    pass
