#!/usr/bin/python
# -*- coding: utf-8 -*-

# See https://getbitbar.com/ for usage


# -- stdlib --
import sys
reload(sys)
sys.setdefaultencoding('utf-8')

# -- third party --
try:
    import requests
except ImportError:
    print('Failed to import `requests`, please install via `sudo pip install requests`')
    sys.exit(1)

# -- own --

# -- code --
# Set this to True if you are authenticating by Kerberos
# Leave it alone if you don't know what Kerberos is
USE_SPNEGO = False

if USE_SPNEGO:
    BASEURL = 'https://DOMAIN'
else:
    BASEURL = 'https://CREDENTIAL@DOMAIN'

SORT_ORDER = {
    'PROBLEM':  'AAAA',
    'FLAPPING': 'BBBB',
    'ACK':      'CCCC',
    'TIMEWAIT': 'ZZZZ',
}
ICONS = {
    "PROBLEM":  "",
    "ACK":      "🔕",
    "FLAPPING": "🎭",  # 🔃  🔄
    "TIMEWAIT": "⌛",
}

if USE_SPNEGO:
    try:
        from requests_gssapi import HTTPSPNEGOAuth
    except ImportError:
        print('Failed to import `requests_gssapi`, please install via `sudo pip install requests_gssapi`')
        sys.exit(1)

    alarms = requests.get(BASEURL + '/alarm/alarms', auth=HTTPSPNEGOAuth()).json()['alarms']
else:
    alarms = requests.get(BASEURL + '/alarm/alarms').json()['alarms']

alarms = [i for i in alarms if i['status'] != 'TIMEWAIT']
alarms.sort(key=lambda a: (SORT_ORDER.get(a['status'], 'ZZZZZ'), a['title']))

if alarms:
    print '😱 = %s' % len(alarms)
    print '---'
    for a in alarms:
        if a['tags']:
            t = ','.join(['%s=%s' % (k, v) for k, v in a['tags'].items()])
            t = '[%s]' % t
        else:
            t = ''

        if USE_SPNEGO:
            print(u'{icon}{a[title]}{t} | bash=/usr/bin/curl param1=-u param2=: param3=--negotiate param4=-XPOST param5={base}/alarm/alarms/{a[id]}/toggle-ack terminal=false refresh=true'.format(
                icon=ICONS.get(a['status'], "❌"),
                a=a, t=t, base=BASEURL,
            ))
        else:
            print(u'{icon}{a[title]}{t} | bash=/usr/bin/curl param1=-XPOST param2={base}/alarm/alarms/{a[id]}/toggle-ack terminal=false refresh=true'.format(
                icon=ICONS.get(a['status'], "❌"),
                a=a, t=t, base=BASEURL,
            ))
else:
    print u'😆'
