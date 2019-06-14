# -*- coding: utf-8 -*-


# -- stdlib --
# -- third party --
import requests

# -- own --
from backend import Backend
from utils import status2emoji


# -- code --
class YunpianSMSBackend(Backend):
    def send(self, ev):
        for user in ev['users']:
            if not user.get('phone'):
                continue

            msg = '【%s】%s[P%s] %s\n' % (
                self.conf['signature'],
                status2emoji(ev['status']),
                ev['level'],
                ev['title'],
            ) + ev['text']

            rst = requests.post('http://yunpian.com/v1/sms/send.json', data={
                'apikey': self.conf['api_key'],
                'mobile': user['phone'],
                'text': msg,
            }).json()

            if rst['code'] != 0:
                raise Exception(rst['detail'])


EXPORT = YunpianSMSBackend
