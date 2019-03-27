# -*- coding: utf-8 -*-

# -- stdlib --

# -- third party --
import requests
from requests.auth import HTTPBasicAuth

# -- own --
from backend import Backend


# -- code --
def status2bccode(s):
    return {
        'PROBLEM': u':scream:',
        'EVENT': u':scream:',
        'OK': u':sweat_smile:',
    }.get(s, s)


class ZulipBackend(Backend):
    def send(self, ev):
        for user in ev['users']:
            if 'zulip' not in user:
                continue

            if 'level' in user:
                # check alarm level only if user defined acceptable level
                if str(ev['level']) not in list(user.get('level')):
                    continue

            title = u'%s[P%s] %s' % (
                status2bccode(ev['status']),
                ev['level'],
                ev['title'],
            )

            channel = str(user.get('channel'))
            topic_prefix = str(user.get('topic'))

            requests.post(
                self.conf['api_url'],
                auth=HTTPBasicAuth(self.conf['username'], self.conf['key']),
                timeout=10,
                data={
                    'type': 'stream',
                    'to': channel,
                    'subject': topic_prefix + '-P' + str(ev['level']),
                    # 'subject': str(self.conf['topic']),
                    'content': title + "\n" + ev['text']
                },
            )


EXPORT = ZulipBackend
