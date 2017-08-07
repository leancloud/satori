# -*- coding: utf-8 -*-

# -- stdlib --
import datetime
import shlex
import re
import logging

# -- third party --
# -- own --

# -- code --
BACKENDS = {}


# from https://stackoverflow.com/questions/1175208/elegant-python-function-to-convert-camelcase-to-snake-case
def get_backend_name(name):
    name = name.replace('Backend', '')
    s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', name)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()


def register_backend(f):
    BACKENDS[get_backend_name(f.__name__)] = f
    return f


def from_string(s):
    return BACKENDS[s]


class Backend(object):
    def __init__(self, conf):
        self.conf = conf
        self.logger = logging.getLogger(self.__class__.__name__)

        self.logger.debug('Initializing backend [{}] ...'.format(self.__class__.__name__))

    def shutdown(self):
        self.logger.debug('Shutting down backend [{}] ...'.format(self.__class__.__name__))

    def send(self, user, event):
        pass