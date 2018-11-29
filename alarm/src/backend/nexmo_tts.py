# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
import nexmo

# -- own --
from backend.common import register_backend, Backend

# -- code --


@register_backend
class NexmoTTSBackend(Backend):
    def send(self, user, event):
        if not user.get('phone'):
            return

        if event['status'] not in ('PROBLEM', 'EVENT'):
            return

        client = nexmo.Client(application_id=self.conf['app_id'], private_key=self.conf['private_key_path'])

        response = client.create_call({
            'to':[{'type':'phone', 'number': str(user['phone'])}],
            'from':{'type':'phone', 'number': str(self.conf['from_number'])},
            'answer_url': self.conf['answer_url']
            })
        self.logger.info('Sending tts phone %s to %s(%s)', response['uuid'], user['name'], user['phone'])