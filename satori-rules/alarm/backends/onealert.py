# -*- coding: utf-8 -*-

# -- stdlib --
import json

# -- third party --
import requests

# -- own --
from backend import Backend


# -- code --
class OneAlertBackend(Backend):
    def send(self, ev):
        api_key = self.conf.get('api_key')
        if not api_key:
            return

        for user in ev['users']:
            resp = requests.post(
                'http://api.110monitor.com/alert/api/event',
                headers={'Content-Type': 'application/json'},
                timeout=10,
                data=json.dumps({
                    "app": user.get('onealert') or api_key,
                    "eventId": ev['id'],
                    "eventType": 'trigger' if ev['status'] in ('PROBLEM', 'EVENT') else 'resolve',
                    "alarmName": ev.get('expr') or ev.get('title', '-'),
                    "entityName": ev['title'],
                    "entityId": ev['title'],
                    "priority": 1,
                    "alarmContent": ev['text'],
                }),
            )

            if not resp.ok:
                raise Exception(resp.json())


EXPORT = OneAlertBackend
