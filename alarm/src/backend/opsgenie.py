# -*- coding: utf-8 -*-

# -- stdlib --
import json
import datetime

# -- third party --
import requests

# -- own --
from backend.common import register_backend, Backend

# -- code --


@register_backend
class OpsgenieBackend(Backend):
    def send(self, users, event):
        for user in users:
            if 'opsgenie_key' not in user:
                continue

            key = user['opsgenie_key']
            if not key:
                continue
            headers={'Content-Type': 'application/json', 'Authorization': 'GenieKey ' + key }

            alarm_id = event['id']  # alias

            # check alerts first
            url = 'https://api.opsgenie.com/v2/alerts/' + alarm_id + '?identifierType=alias'
            resp = None
            # https://docs.opsgenie.com/docs/alert-notifications-flow
            alert_status = None
            ack = None
            snoozed = None
            # https://docs.opsgenie.com/docs/alert-api#section-get-alert
            try:
                resp = requests.get( url, headers=headers, timeout=10 )
                if resp.status_code == 200:
                    alert_status = resp.json()['data']['status']
                    ack = resp.json()['data']['acknowledged']
                    snoozed = resp.json()['data']['snoozed']
            except:
                pass

            message = event['title']
            description = event['text']
            priority = 'P' + str(event['level']+1)
            details = event['tags']  # actually is TAGS

            # fill teams
            teams = []
            for team in event['groups']:
                teams.append( { 'name': team, 'type': 'team'})

            url = ''
            body = {}
            # status: PROBLEM OK EVENT FLAPPING TIMEWAIT ACK
            if event['status'] in ('PROBLEM', 'EVENT', 'FLAPPING'):
                # 'CRITICAL'
                if alert_status == 'open' and ack == True :
                    continue
                url = 'https://api.opsgenie.com/v2/alerts' 
                body = {'message':message, 'alias': alarm_id, 'description':description, 
                        'priority':priority, 'responders': teams, 'details':details,
                        'note': event['note'] }
            elif event['status'] in ( 'OK' ):
                # 'RECOVERY'
                if alert_status == 'closed':
                    continue
                url = 'https://api.opsgenie.com/v2/alerts/' + alarm_id + '/close?identifierType=alias'
                body = { 'user':'satori', 'source': 'satori-backend', 'note':'(mark or auto) recovery from satori'}
            elif event['status'] in ( 'TIMEWAIT'):
                # 'SNOOZE'
                if snoozed == True:
                    continue
                endtime = datetime.datetime.now() + datetime.timedelta(minutes = 10)
                url = 'https://api.opsgenie.com/v2/alerts/' + alarm_id + '/snooze?identifierType=alias'
                body = { 'user':'satori', 'source': 'satori-backend', 
                        'note':'(mark or auto) recovery from satori', 'end': endtime.isoformat() }
            elif event['status'] == 'ACK':
                # 'ACK'
                if alert_status == 'open' and ack == True :
                    continue
                url = 'https://api.opsgenie.com/v2/alerts/' + alarm_id + '/acknowledge?identifierType=alias'
                body = { 'user':'satori', 'source': 'satori-backend', 'note':'ack from satori'}
            else:
                # 'INFO' , Low Priority
                url = 'https://api.opsgenie.com/v2/alerts' 
                priority = 'P5'

            try:
                resp = None
                resp = requests.post( url, headers=headers, timeout=10, data=json.dumps( body ))
                self.logger.info( 'notify opsgenie %s, %s, %s', alarm_id, resp.json(), event)
            except:
                if resp:
                    self.logger.error( 'notify opsgenie failed: %s, %s, %s, %s', resp.text, url, headers, body )
                else:
                    self.logger.error( 'notify opsgenie failed: %s, %s, %s', url, headers, body )
                raise
