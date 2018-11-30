# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- stdlib --
import json

# -- third party --
from gevent.lock import RLock
import gevent
import requests

# -- own --
from backend.common import register_backend, Backend
from utils import status2emoji


# -- code --
'''
  - name: wechat
    backend: wechat_qy
    level: '0123'
    corpid: xxx
    secret: yyy
    agentid: 1

  user:
    name: "foo"
    email: foo@leancloud.rocks
    phone: 18911674224
    wechat: foo
    wechat_party: bar
'''

import time


@register_backend
class WechatQYBackend(Backend):
    def __init__(self, conf):
        super(WechatQYBackend, self).__init__(conf)

        self.lock = RLock()
        self.token_cache = {}

    def get_access_token(self, corp_id, secret):
        key = '%s:%s' % (corp_id, secret)

        def get():
            token, expire = self.token_cache.get(key, (None, 0))

            if time.time() - expire >= 7100:
                token = None

            return token

        token = get()
        if token:
            return token

        with self.lock:
            token = get()
            if token:
                return token

            for _ in xrange(3):
                resp = requests.get('https://qyapi.weixin.qq.com/cgi-bin/gettoken',
                    params={'corpid': corp_id, 'corpsecret': secret},
                    timeout=10,
                )

                if not resp.ok:
                    raise Exception('Error getting access token: %s' % resp.content)

                ret = resp.json()

                code = ret.get('errcode', 0)
                if code == 42001:
                    gevent.sleep(2)
                    continue
                elif code:
                    raise Exception('Error getting access token: %s' % ret['errmsg'])

                token = ret['access_token']
                break
            else:
                self.logger.error('Get access token max retry exceeded')
                return None

            self.token_cache[key] = (token, time.time())
            return token

    def clear_access_token(self, corp_id, secret):
        key = '%s:%s' % (corp_id, secret)
        self.token_cache.pop(key)

    def send(self, users, event):
        for user in users:
            corp_id  = user.get('wechat_corpid') or self.conf['corpid']
            secret   = user.get('wechat_secret') or self.conf['secret']
            agent_id = user.get('wechat_agentid') or self.conf['agentid']

            touser = user.get('wechat', '')
            toparty = user.get('wechat_party', '')

            if not (touser or toparty):
                continue

            msg = u'%s[P%s] %s\n' % (
                status2emoji(event['status']),
                event['level'],
                event['title'],
            ) + event['text']

            payload = {
                "touser": touser,
                "toparty": toparty,
                "msgtype": "text",
                "agentid": agent_id,
                "text": {
                    "content": msg,
                },
                "safe": 0,
            }

            self.logger.info('Sending wechat message to %s %s', touser, toparty)

            for _ in xrange(2):
                token = self.get_access_token(corp_id, secret)
                if not token:
                    return

                resp = requests.post(
                    'https://qyapi.weixin.qq.com/cgi-bin/message/send',
                    params={'access_token': token},
                    headers={'Content-Type': 'application/json'},
                    timeout=10,
                    data=json.dumps(payload).decode('unicode-escape').encode('utf-8'),
                )

                if not resp.ok:
                    raise Exception(resp.content)

                ret = resp.json()
                if ret['errcode'] == 40014:
                    self.clear_access_token(corp_id, secret)
                    continue
                elif ret['errcode'] != 0:
                    raise Exception(resp.content)

                break
            else:
                raise Exception('Too many retries')