# -*- coding: utf-8 -*-

# -- stdlib --
import json

# -- third party --
import requests

# -- own --
from backend import Backend


# -- code --
class VictorOpsBackend(Backend):
    def send(self, ev):
        for user in ev['users']:
            if 'victorops' not in user:
                continue

            url = user['victorops']
            try:
                routing_key = ev['groups'][0]
            except:
                routing_key = 'default'

            # status: PROBLEM OK EVENT FLAPPING TIMEWAIT ACK
            if ev['status'] in ('PROBLEM', 'EVENT'):
                msg_type = 'CRITICAL'
            elif ev['status'] in ('OK', 'TIMEWAIT'):
                msg_type = 'RECOVERY'
            elif ev['status'] == 'ACK':
                msg_type = 'ACK'
            else:
                msg_type = 'INFO'

            resp = requests.post(
                url + '/' + routing_key,
                headers={'Content-Type': 'application/json'},
                timeout=10,
                data=json.dumps({
                    'entity_id': ev['title'],
                    'entity_display_name': ev['title'],
                    'priority': ev['level'],
                    'message_type': msg_type,
                    'state_message': ev['text'],
                }),
            )
            if not resp.ok:
                raise Exception(resp.text)


EXPORT = VictorOpsBackend
