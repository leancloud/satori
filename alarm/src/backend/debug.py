# -*- coding: utf-8 -*-
# -- stdlib --
# -- third party --
# -- own --
from backend.common import register_backend

# -- code --


@register_backend
def debug(conf, user, event):
    import pprint
    print '>>>' + '=' * 77
    pprint.pprint(user)
    print '-' * 80
    pprint.pprint(event)
    print '<<<' + '=' * 77
