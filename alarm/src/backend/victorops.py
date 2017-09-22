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
        routing_key = event['related_groups'][0]

        #PROBLEM OK EVENT FLAPPING TIMEWAIT ACK
        if event['status'] in ('PROBLEM', 'EVENT', 'FLAPPING'):
            msg_type = 'CRITICAL'
        elif event['status'] in ( 'OK', 'TIMEWAIT'):
            msg_type = 'RECOVERY'
        else:
            msg_type = 'ACK'
        # for leancloud only
        if event['level'] == 5:
            msg_type = 'INFO'

        resp = requests.post(
            url + '/' + routing_key,
            headers={'Content-Type': 'application/json'},
            timeout=10,
            data=json.dumps({
                'entity_id': event['title'],
                'entity_display_name': event['title'],
                'message_type': msg_type,
                'state_message': event['text'],
            }),
        )
        if not resp.ok:
            raise Exception(resp.json())
