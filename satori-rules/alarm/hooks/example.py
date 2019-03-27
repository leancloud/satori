# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
# -- own --
from hook import Hook


# -- code --
class ExampleHook(Hook):
    sort = 0

    def call(self, ev_type, arg):
        if ev_type == 'send_alarm':
            ev, backend = arg
            # do some interesting stuff
            # self.logger.debug('Hook event: %s, %s, %s, %s', ev_type, ev, backend)
            return arg

        return arg


EXPORT = ExampleHook
