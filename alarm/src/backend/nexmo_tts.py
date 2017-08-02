# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
import requests

# -- own --
from backend.common import register_backend, Backend

# -- code --


@register_backend
class NexmoTTSBackend(Backend):
    def send(self, user, event):
        if not user.get('phone'):
            return

        if event['status'] not in ('PROBLEM', 'EVENT'):
            return

        requests.post('https://api.nexmo.com/tts/json', params={
            'api_key': str(self.conf['api_key']),
            'api_secret': str(self.conf['api_secret']),
            'to': '86' + str(user['phone']),
            'voice': self.conf['voice'],
            'lg': self.conf['lg'],
            'repeat': self.conf['repeat'],
        }, data={'text': self.conf['prefix'] + event['note']})
