# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
import nexmo
import time

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
                phones.append( '86' + str(user.get('phone')))
        else:
            if not user.get('phone'):
                return
            phones.append( '86' + str(user.get('phone')))

        client = nexmo.Client(application_id=self.conf['app_id'], private_key=self.conf['private_key_path'])
        for phone in phones:
            try:
                # Make sure not reach request api limit
                time.sleep(1)
                response = client.create_call({
                    'to': [{'type':'phone','number': phone }],
                    'from':{'type':'phone', 'number': str(self.conf['from_number'])},
                    'answer_url': self.conf['answer_url']
                    })
                self.logger.info( 'Phone call %s done, uuid: %s', str(phone), response['uuid'] )
            except:
                if response:
                    self.logger.error( 'Phone call %s failed, uuid: %s', str(phone), response['uuid'] )
                else:
                    raise
