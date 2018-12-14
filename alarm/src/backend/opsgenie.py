# -*- coding: utf-8 -*-

# -- stdlib --
import json

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
                return

            satori_id = event['id']  # alias
            message = event['title']
            description = event['text']
            priority = 'P' + str(event['level'])
            details = event['tags']  # actually is TAGS

            # fill teams
            teams = []
            for team in event['groups']:
                teams.append( { 'name': team, 'type': 'team'})

            url = ''
            body = {}
            headers={'Content-Type': 'application/json', 'Authorization': 'GenieKey ' + key },
            # status: PROBLEM OK EVENT FLAPPING TIMEWAIT ACK
            if event['status'] in ('PROBLEM', 'EVENT', 'FLAPPING'):
                # 'CRITICAL'
                url = 'https://api.opsgenie.com/v2/alerts' 
                body = {'message':message, 'alias': satori_id, 'description':description, 
                        'priority':priority, 'responders': teams, 'details':details,
                        'note': event['note'] }
            elif event['status'] in ( 'OK', 'TIMEWAIT'):
                # 'RECOVERY'
                url = 'https://api.opsgenie.com/v2/alerts/' + satori_id + '/close?identifierType=alias'
                body = { 'user':'satori', 'source': 'satori-backend', 'note':'(mark or auto) recovery from satori'}
            elif event['status'] == 'ACK':
                # 'ACK'
                url = 'https://api.opsgenie.com/v2/alerts/' + satori_id + '/acknowledge?identifierType=alias'
                body = { 'user':'satori', 'source': 'satori-backend', 'note':'ack from satori'}
            else:
                # 'INFO' , Low Priority
                url = 'https://api.opsgenie.com/v2/alerts' 
                priority = 'P5'

            try:
                resp = None
                resp = requests.post( url, headers=headers, timeout=10, data=json.dumps( body ))
                self.logger.info( 'notify opsgenie %s, %s', satori_id, event)
            except:
                if resp:
                    self.logger.error( 'notify opsgenie failed: %s, %s, %s, %s', resp.text, url, headers, body )
                else:
                    self.logger.error( 'notify opsgenie failed: %s, %s, %s', url, headers, body )
                raise
