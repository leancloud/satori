# -*- coding: utf-8 -*-
# -- stdlib --
# -- third party --
# -- own --
from backend.common import register_backend, Backend

# -- code --


@register_backend
class DebugBackend(Backend):
    def send(self, user, event):
        import pprint
        print '>>>' + '=' * 77
        pprint.pprint(user)
        print '-' * 80
        pprint.pprint(event)
        print '<<<' + '=' * 77
