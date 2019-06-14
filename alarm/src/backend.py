# -*- coding: utf-8 -*-

# -- stdlib --
import importlib
import logging
import os
import os.path
import sys

# -- third party --
# -- own --

# -- code --
BACKENDS = {}
log = logging.getLogger('backend')


def from_string(s):
    return BACKENDS[s]


def scrape_backends(path):
    path = os.path.abspath(os.path.join(path, 'alarm'))

    if path not in sys.path:
        sys.path.append(path)

    BACKENDS.clear()

    for dirpath, dirnames, filenames in os.walk(os.path.join(path, 'backends')):
        for fn in filenames:
            if not fn.endswith('.py'):
                continue

            if fn == '__init__.py':
                continue

            modname = fn[:-3]
            modfullname = 'backends.%s' % modname
            sys.modules.pop(modfullname, '')

            try:
                mod = importlib.import_module(modfullname)
                if not hasattr(mod, 'EXPORT'):
                    log.error('Backend module `%s` not exporting a backend', modname)
                    continue

                BACKENDS[modname] = mod.EXPORT
                log.info('Loaded backend `%s`', modname)

            except Exception:
                log.exception('Error importing `%s`', modname)

        break


class Backend(object):
    def __init__(self, conf):
        self.conf = conf
        self.logger = logging.getLogger(self.__class__.__name__)

        self.logger.debug('Initializing backend [{}] ...'.format(self.__class__.__name__))

    def shutdown(self):
        self.logger.debug('Shutting down backend [{}] ...'.format(self.__class__.__name__))

    def send(self, ev):
        pass
