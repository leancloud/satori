# -*- coding: utf-8 -*-

# -- stdlib --
import json

# -- third party --
import requests
from requests.auth import HTTPBasicAuth

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
class ZulipBackend(Backend):
    def send(self, user, event):
        if 'zulip' not in user:
            return

        title = u'%s[P%s] %s' % (
            status2bccode(event['status']),
            event['level'],
            event['title'],
        )

        requests.post(
            self.conf['api_url'],
            auth=HTTPBasicAuth( self.conf['username'], self.conf['key']),
            timeout=10,
            data={  'type': 'stream',
                    'to': str(self.conf['group']),
                    'subject': str(self.conf['channel']),
                    'content': title + "\n" + event['text'] },
        )