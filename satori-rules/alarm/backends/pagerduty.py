# -*- coding: utf-8 -*-

# -- stdlib --
import json

# -- third party --
import requests

# -- own --
from backend import Backend


# -- code --
class PagerDutyBackend(Backend):
    def send(self, ev):
        api_key = self.conf.get('api_key')
        if not api_key:
            return

        for user in ev['users']:
            if 'pagerduty' in user:
                api_key = user['pagerduty']

            resp = requests.post(
                'https://events.pagerduty.com/generic/2010-04-15/create_event.json',
                headers={'Content-Type': 'application/json'},
                timeout=10,
                data=json.dumps({
                    'service_key': api_key,
                    'incident_key': ev['id'],
                    'event_type': 'trigger' if ev['status'] in ('PROBLEM', 'EVENT') else 'resolve',
                    'description': ev['title'],
                    'details': {'detail': ev['text']},
                }),
            )
            if not resp.ok:
                raise Exception(resp.json())


EXPORT = PagerDutyBackend
