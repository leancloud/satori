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
        api_key = self.conf.get('api_key')

        if 'victorops' not in user:
            return

        url = user['victorops_url']
        if not url:
            return

        resp = requests.post(
            url,
            headers={'Content-Type': 'application/json'},
            timeout=10,
            data=json.dumps({
                'entity_id': event['title'],
                'entity_display_name': event['title'],
                'message_type': 'CRITICAL' if event['status'] in ('PROBLEM', 'EVENT') else 'RECOVERY',
                'state_message': event['text'],
            }),
        )
        if not resp.ok:
            raise Exception(resp.json())
