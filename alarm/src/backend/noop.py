# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
# -- own --
from backend.common import register_backend, Backend

# -- code --


@register_backend
class NoopBackend(Backend):
    pass

