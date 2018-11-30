# -*- coding: utf-8 -*-

# -- stdlib --
import json

# -- third party --
import requests

# -- own --
from utils import status2emoji
from backend.common import register_backend, Backend

# -- code --

def status2bccode(s):
    return {
        'PROBLEM': u':scream:',
        'EVENT': u':scream:',
        'OK': u':sweat_smile:',
    }.get(s, s)


@register_backend
class BearychatBackend(Backend):
    def send(self, users, event):
        for user in users:
            if 'bearychat' not in user:
                continue

            url = user['bearychat']

            if event['status'] in ('PROBLEM', 'EVENT'):
                color = [
                    u'#be10c2',  # purple 0
                    u'#ef1000',  # red 1
                    u'#fbb726',  # orange 2
                    u'#fdfd00',  # yellow 3
                    u'#f5f5f5',  # grey 4+
                ][min(event['level'], 4)]
            else:
                color = u'#5cab2a'  # green

            title = u'%s[P%s] %s' % (
                status2bccode(event['status']),
                event['level'],
                event['title'],
            )
            requests.post(
                url,
                headers={'Content-Type': 'application/json'},
                timeout=10,
                data=json.dumps({
                    'text': title,
                    'attachments': [{
                        'title': event['status'],
                        'text': event['text'],
                        'color': color,
                    }],
                }),
            )