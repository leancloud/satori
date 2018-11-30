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
    def send(self, users, event):
        for user in users:
            if 'zulip' not in user:
                continue

            title = u'%s[P%s] %s' % (
                status2bccode(event['status']),
                event['level'],
                event['title'],
            )

            try:
                response = requests.post(
                    self.conf['api_url'],
                    auth=HTTPBasicAuth( self.conf['username'], self.conf['key']),
                    timeout=10,
                    data={  'type': 'stream',
                            'to': str(self.conf['channel']),
                            'subject': str(self.conf['topic'])+'-P'+str(event['level']),
                            #'subject': str(self.conf['topic']),
                            'content': title + "\n" + event['text'] },
                )
            except:
                self.logger.info('notiy zulip failed: %s' % response.text)
                raise