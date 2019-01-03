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

            if 'level' in user:
                # check alarm level only if user defined acceptable level
                if str(event['level']) not in list(user.get('level')):
                    continue

            title = u'%s[P%s] %s' % (
                status2bccode(event['status']),
                event['level'],
                event['title'],
            )

            channel = str(user.get('channel'))
            topic_prefix = str(user.get('topic'))
            try:
                response = requests.post(
                    self.conf['api_url'],
                    auth=HTTPBasicAuth( self.conf['username'], self.conf['key']),
                    timeout=10,
                    data={  'type': 'stream',
                            'to': channel,
                            'subject': topic_prefix + '-P' + str(event['level']),
                            #'subject': str(self.conf['topic']),
                            'content': title + "\n" + event['text'] },
                )
            except:
                raise