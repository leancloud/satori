# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
# -- own --
from backend import Backend


# -- code --
class DebugBackend(Backend):
    def send(self, ev):
        for user in ev['users']:
            import pprint
            print '>>>' + '=' * 77
            pprint.pprint(user)
            print '-' * 80
            pprint.pprint(ev)
            print '<<<' + '=' * 77


EXPORT = DebugBackend
