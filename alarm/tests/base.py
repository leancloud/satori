# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- stdlib --
# -- third party --
import yaml

# -- own --

# -- code --


class TestBase(object):
    @classmethod
    def setUpClass(cls):
        from state import State
        from main import AlarmDFA
        import config
        confs = yaml.load(open('tests/config.yaml').read())
        State.config = confs
        State.alarms = AlarmDFA()
        State.options = {}
        State.userconf = {}

        config.refresh()
