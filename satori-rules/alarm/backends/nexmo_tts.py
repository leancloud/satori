# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
import nexmo
import time

# -- own --
from backend import Backend


# -- code --
class NexmoTTSBackend(Backend):
    def send(self, ev):
        if ev['status'] not in ('PROBLEM', 'EVENT'):
            return

        phones = []
        for user in ev['users']:
            if not user.get('phone'):
                continue
            phones.append('86' + str(user.get('phone')))

        client = nexmo.Client(
            application_id=self.conf['app_id'],
            private_key=self.conf['private_key_path'],
        )

        for phone in phones:
            try:
                # Make sure not reach request api limit
                time.sleep(1)
                response = client.create_call({
                    'to': [{
                        'type': 'phone',
                        'number': phone
                    }],
                    'from': {
                        'type': 'phone',
                        'number': str(self.conf['from_number'])
                    },
                    'answer_url':
                    self.conf['answer_url']
                })
                self.logger.info('Phone call %s done, uuid: %s', str(phone), response['uuid'])
            except Exception:
                self.logger.exception('Phone call %s failed, uuid: %s', str(phone))


EXPORT = NexmoTTSBackend
