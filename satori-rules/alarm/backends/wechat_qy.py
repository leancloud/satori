# -*- coding: utf-8 -*-


# -- stdlib --
import time
import json
import codecs

# -- third party --
from gevent.lock import RLock
import gevent
import requests

# -- own --
from backend import Backend
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

            for _ in range(3):
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

    def send(self, ev):
        for user in ev['users']:
            corp_id  = user.get('wechat_corpid') or self.conf['corpid']
            secret   = user.get('wechat_secret') or self.conf['secret']
            agent_id = user.get('wechat_agentid') or self.conf['agentid']

            touser = user.get('wechat', '')
            toparty = user.get('wechat_party', '')

            if not (touser or toparty):
                continue

            text = ev['text']
            text = text if len(text) < 500 else "[消息太长了，无法从微信发出]"
            msg = '%s[P%s] %s\n' % (
                status2emoji(ev['status']),
                ev['level'],
                ev['title'],
            ) + text

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

            for _ in range(2):
                token = self.get_access_token(corp_id, secret)
                if not token:
                    return

                resp = requests.post(
                    'https://qyapi.weixin.qq.com/cgi-bin/message/send',
                    params={'access_token': token},
                    headers={'Content-Type': 'application/json'},
                    timeout=10,
                    data=json.dumps(payload),
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


EXPORT = WechatQYBackend
