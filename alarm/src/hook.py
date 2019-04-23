# -*- coding: utf-8 -*-

# -- stdlib --
import os
import sys
import importlib
import logging
import os.path

# -- third party --
# -- own --

# -- code --
HOOKS = []
log = logging.getLogger('hook')


def scrape_hooks(path):
    path = os.path.abspath(os.path.join(path, 'alarm'))

    if path not in sys.path:
        sys.path.append(path)

    HOOKS[:] = []

    for dirpath, dirnames, filenames in os.walk(os.path.join(path, 'hooks')):
        for fn in filenames:
            if not fn.endswith('.py'):
                continue

            if fn == '__init__.py':
                continue

            modname = fn[:-3]
            modfullname = 'hooks.%s' % modname
            sys.modules.pop(modfullname, '')

            try:
                mod = importlib.import_module(modfullname)
                if not hasattr(mod, 'EXPORT'):
                    log.error('Hook module `%s` not exporting a hook', modname)
                    continue

                HOOKS.append(mod.EXPORT())
                log.info('Loaded hook `%s`', modname)

            except Exception:
                log.exception('Error importing `%s`', modname)
                raise

        HOOKS.sort(key=lambda hook: hook.sort)

        break


def call_hooks(ev_type, arg):
    log.debug('Calling `%s` hooks for %s', ev_type, arg)

    for hook in HOOKS:
        arg = hook.call(ev_type, arg)

    return arg


class Hook(object):
    sort = 0

    def __init__(self):
        self.logger = logging.getLogger(self.__class__.__name__)
        self.logger.debug('Initializing hook [{}] ...'.format(self.__class__.__name__))

    def call(self, ev_type, arg):
        if ev_type == 'send_alarm':
            ev, user, backend = arg
            # do some interesting stuff
            return arg

        return arg
