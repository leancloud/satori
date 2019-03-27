# -*- coding: utf-8 -*-

# -- stdlib --
import json

# -- third party --
import requests

# -- own --
from backend import Backend


# -- code --
def status2bccode(s):
    return {
        'PROBLEM': u':scream:',
        'EVENT': u':scream:',
        'OK': u':sweat_smile:',
    }.get(s, s)


class BearychatBackend(Backend):
    def send(self, ev):
        for user in ev['users']:
            if 'bearychat' not in user:
                continue

            url = user['bearychat']

            if ev['status'] in ('PROBLEM', 'EVENT'):
                color = [
                    u'#be10c2',  # purple 0
                    u'#ef1000',  # red 1
                    u'#fbb726',  # orange 2
                    u'#fdfd00',  # yellow 3
                    u'#f5f5f5',  # grey 4+
                ][min(ev['level'], 4)]
            else:
                color = u'#5cab2a'  # green

            title = u'%s[P%s] %s' % (
                status2bccode(ev['status']),
                ev['level'],
                ev['title'],
            )
            requests.post(
                url,
                headers={'Content-Type': 'application/json'},
                timeout=10,
                data=json.dumps({
                    'text': title,
                    'attachments': [{
                        'title': ev['status'],
                        'text': ev['text'],
                        'color': color,
                    }],
                }),
            )


EXPORT = BearychatBackend
