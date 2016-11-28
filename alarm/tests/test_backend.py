# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- prioritized --
from gevent import monkey
monkey.patch_all()

# -- stdlib --
# -- third party --
# -- own --
from base import TestBase
from example_data import alarm_example, user_example


# -- code --
class TestBackend(TestBase):

    def do_tezt_backend(self, name):
        import main
        import backend
        from state import State
        confs = {i['backend']: i for i in State.strategies.values()}
        f = backend.from_string(name)
        f(confs[name], user_example, main.cook_event(alarm_example))

    # def test_bearychat(self):
    #    return self.do_tezt_backend('bearychat')

    # def test_nexmo_tts(self):
    #     return self.do_tezt_backend('nexmo_tts')

    def test_noop(self):
        return self.do_tezt_backend('noop')

    # def test_pagerduty(self):
    #     return self.do_tezt_backend('pagerduty')

    # def test_smtp(self):
    #     return self.do_tezt_backend('smtp')

    # def test_yunpian_sms(self):
    #     return self.do_tezt_backend('yunpian_sms')

    # def test_onealert(self):
    #    return self.do_tezt_backend('onealert')

    def test_wechat_qy(self):
        return self.do_tezt_backend('wechat_qy')
