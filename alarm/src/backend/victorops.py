# -*- coding: utf-8 -*-

# -- stdlib --
import json

# -- third party --
import requests

# -- own --
from backend.common import register_backend, Backend

# -- code --


@register_backend
class VictorOpsBackend(Backend):
    def send(self, user, event):
        if 'victorops' not in user:
            return

        url = user['victorops']
        try:
            routing_key = event['groups'][0]
        except:
            routing_key = 'default'

        # status: PROBLEM OK EVENT FLAPPING TIMEWAIT ACK
        if event['status'] in ('PROBLEM', 'EVENT'):
            msg_type = 'CRITICAL'
        elif event['status'] in ( 'OK', 'TIMEWAIT'):
            msg_type = 'RECOVERY'
        elif event['status'] == 'ACK':
            msg_type = 'ACK'
        else:
            msg_type = 'INFO'

        resp = requests.post(
            url + '/' + routing_key,
            headers={'Content-Type': 'application/json'},
            timeout=10,
            data=json.dumps({
                'entity_id': event['title'],
                'entity_display_name': event['title'],
                'priority': event['level'],
                'message_type': msg_type,
                'state_message': event['text'],
            }),
        )
        if not resp.ok:
            raise Exception(resp.text)
