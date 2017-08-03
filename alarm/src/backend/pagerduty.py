# -*- coding: utf-8 -*-

# -- stdlib --
import json

# -- third party --
import requests

# -- own --
from backend.common import register_backend, Backend

# -- code --


@register_backend
class PagerDutyBackend(Backend):
    def send(self, user, event):
        api_key = self.conf.get('api_key')

        if 'pagerduty' in user:
            api_key = user['pagerduty']

        if not api_key:
            return

        resp = requests.post(
            'https://events.pagerduty.com/generic/2010-04-15/create_event.json',
            headers={'Content-Type': 'application/json'},
            timeout=10,
            data=json.dumps({
                'service_key': api_key,
                'incident_key': event['id'],
                'event_type': 'trigger' if event['status'] in ('PROBLEM', 'EVENT') else 'resolve',
                'description': event['title'],
                'details': {'detail': event['text']},
            }),
        )
        if not resp.ok:
            raise Exception(resp.json())
