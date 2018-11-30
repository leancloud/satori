# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
import nexmo

# -- own --
from backend.common import register_backend, Backend

# -- code --


@register_backend
class NexmoTTSBackend(Backend):
    def send(self, users, event):
        if event['status'] not in ('PROBLEM', 'EVENT'):
            return

        phones = []
        if isinstance( users, list):
            for user in users:
                if not user.get('phone'):
                    continue
                phones.append( { 'type':'phone', 'number': '86' + str(user.get('phone'))})
        else:
            if not user.get('phone'):
                return
            phones.append( { 'type':'phone', 'number': '86' + str(users.get('phone'))})

        client = nexmo.Client(application_id=self.conf['app_id'], private_key=self.conf['private_key_path'])
        try:
            response = client.create_call({
                'to': phones,
                'from':{'type':'phone', 'number': str(self.conf['from_number'])},
                'answer_url': self.conf['answer_url']
                })
        except:
            self.logger.info( phones )
            raise
        self.logger.info('Sending tts phone %s to %s', response['uuid'], str(phones))