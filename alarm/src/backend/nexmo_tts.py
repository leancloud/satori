# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
import requests

# -- own --
from backend.common import register_backend

# -- code --


@register_backend
def nexmo_tts(conf, user, event):
    if not user.get('phone'):
        return

    if event['status'] not in ('PROBLEM', 'EVENT'):
        return

    requests.post('https://api.nexmo.com/tts/json', params={
        'api_key': str(conf['api_key']),
        'api_secret': str(conf['api_secret']),
        'to': '86' + str(user['phone']),
        'voice': conf['voice'],
        'lg': conf['lg'],
        'repeat': conf['repeat'],
    }, data={'text': conf['prefix'] + event['note']})
