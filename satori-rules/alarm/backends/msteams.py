# -*- coding: utf-8 -*-

# -- stdlib --
import json

# -- third party --
import requests

# -- own --
from backend import Backend


# -- code --


class MicrososftTeamsBackend(Backend):
    def send(self, ev):
        if ev['status'] in ('PROBLEM', 'EVENT'):
            color = [
                'BE10C2',  # purple 0
                'EF1000',  # red 1
                'FBB726',  # orange 2
                'FDFD00',  # yellow 3
                'F5F5F5',  # grey 4+
            ][min(ev['level'], 4)]
        else:
            color = '#5cab2a'  # green

        title = '%s[P%s] %s' % (
            '😱' if ev['status'] in ('PROBLEM', 'EVENT') else '😅',
            ev['level'],
            ev['title'],
        )
        facts = [
            ('Time', ev['formatted_time']),
            ('Metric', ev['metric']),
            ('Value', str(ev['actual'])),
        ]
        facts.extend([
            ('Tag: %s' % t, str(v))
            for t, v in ev['tags'].items()
        ])
        facts = [{'name': n, 'value': v} for n, v in facts]

        payload = {
            "@type": "MessageCard",
            "@context": "http://schema.org/extensions",
            "themeColor": color,
            "summary": title,
            "sections": [{
                "activityTitle": title,
                "activitySubtitle": ev['status'],
                "facts": facts,
            }],
        }

        if ev['description']:
            payload['sections'].append({'text': "```\n%s\n```" % ev['description'], 'markdown': True})

        # TODO: support MessageCard actions

        for user in ev['users']:
            if 'msteams' not in user:
                continue

            url = user['msteams']

            requests.post(
                url,
                headers={'Content-Type': 'application/json'},
                timeout=10,
                data=json.dumps(payload),
            )


EXPORT = MicrososftTeamsBackend
