# -*- coding: utf-8 -*-

# -- stdlib --
import json
import datetime

# -- third party --
import requests

# -- own --
from backend import Backend


# -- code --
class OpsgenieBackend(Backend):
    def send(self, ev):
        for user in ev['users']:
            key = user.get('opsgenie_key')
            if not key:
                continue

            headers = {
                'Content-Type': 'application/json',
                'Authorization': 'GenieKey ' + key
            }

            alarm_id = ev['id']  # alias

            # check alerts first
            url = 'https://api.opsgenie.com/v2/alerts/%s?identifierType=alias' % alarm_id

            resp = None
            # https://docs.opsgenie.com/docs/alert-notifications-flow
            alert_status = None
            ack = None
            snoozed = None
            # https://docs.opsgenie.com/docs/alert-api#section-get-alert
            try:
                resp = requests.get(url, headers=headers, timeout=10)
                if resp.status_code == 200:
                    alert_status = resp.json()['data']['status']
                    ack = resp.json()['data']['acknowledged']
                    snoozed = resp.json()['data']['snoozed']
            except Exception:
                self.logger.exception('Error querying existing alarms, continuing anyway.')

            message = ev['title']
            description = ev['text']
            priority = 'P' + str(ev['level'] + 1)
            details = ev['tags']  # actually is TAGS

            # fill teams
            teams = []
            for team in ev['groups']:
                teams.append({'name': team, 'type': 'team'})

            url = ''
            body = {}
            # status: PROBLEM OK EVENT FLAPPING TIMEWAIT ACK
            if ev['status'] in ('PROBLEM', 'EVENT', 'FLAPPING'):
                # 'CRITICAL'
                if alert_status == 'open' and ack:
                    continue
                url = 'https://api.opsgenie.com/v2/alerts'
                body = {
                    'message': message,
                    'alias': alarm_id,
                    'description': description,
                    'priority': priority,
                    'responders': teams,
                    'details': details,
                    'note': ev['note']
                }

            elif ev['status'] == 'OK':
                # 'RECOVERY'
                if alert_status == 'closed':
                    continue

                url = 'https://api.opsgenie.com/v2/alerts/%s/close?identifierType=alias' % alarm_id
                body = {
                    'user': 'satori',
                    'source': 'satori-backend',
                    'note': '(mark or auto) recovery from satori'
                }

            elif ev['status'] == 'TIMEWAIT':
                # 'SNOOZE'
                if snoozed:
                    continue

                endtime = datetime.datetime.now() + datetime.timedelta(minutes=10)
                url = 'https://api.opsgenie.com/v2/alerts/%s/snooze?identifierType=alias' % alarm_id
                body = {
                    'user': 'satori',
                    'source': 'satori-backend',
                    'note': '(mark or auto) recovery from satori',
                    'end': endtime.isoformat()
                }

            elif ev['status'] == 'ACK':
                # 'ACK'
                if alert_status == 'open' and ack:
                    continue

                url = 'https://api.opsgenie.com/v2/alerts/%s/acknowledge?identifierType=alias' % alarm_id
                body = {
                    'user': 'satori',
                    'source': 'satori-backend',
                    'note': 'ack from satori'
                }

            else:
                # 'INFO' , Low Priority
                url = 'https://api.opsgenie.com/v2/alerts'
                priority = 'P5'

            try:
                resp = None
                resp = requests.post(url, headers=headers, timeout=10, data=json.dumps(body))
                self.logger.info('notify opsgenie %s, %s, %s', alarm_id, resp.json(), ev)
            except Exception:
                self.logger.exception('OpsGenie API failed')

                if resp:
                    self.logger.error(
                        'notify opsgenie failed: %s, %s, %s, %s',
                        resp.text, url, headers, body,
                    )
                else:
                    self.logger.error(
                        'notify opsgenie failed: %s, %s, %s',
                        url, headers, body,
                    )


EXPORT = OpsgenieBackend
