#!/usr/bin/python
# -*- coding: utf-8 -*-

# See https://getbitbar.com/ for usage


# -- stdlib --
import sys
reload(sys)
sys.setdefaultencoding('utf-8')

# -- third party --
import requests

# -- own --

# -- code --

BASEURL = 'http://YOUR_URL'
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

        print u'{icon}{a[title]}{t} | bash=/usr/bin/curl param1=-XPOST param2={base}/alarm/alarms/{a[id]}/toggle-ack terminal=false refresh=true'.format(
            icon=ICONS.get(a['status'], "❌"),
            a=a, t=t, base=BASEURL,
        )
else:
    print u'😆'
