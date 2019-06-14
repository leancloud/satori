# -*- coding: utf-8 -*-
from __future__ import absolute_import, division, print_function, unicode_literals

# -- stdlib --
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from email.utils import formatdate
import smtplib

# -- third party --
# -- own --
from backend import Backend
from utils import status2emoji


# -- code --
def send_mail(send_from, send_to, subject, text, files=[], server="localhost", ssl=False, username=None, password=None):
    msg = MIMEMultipart('alternative')
    msg.set_charset('utf-8')
    msg['From'] = send_from
    msg['To'] = send_to
    msg['Date'] = formatdate(localtime=True)
    msg['Subject'] = subject
    part = MIMEText(text)
    part.set_charset('utf-8')
    msg.attach(part)
    if ssl:
        smtp = smtplib.SMTP_SSL(server)
    else:
        smtp = smtplib.SMTP(server)
    if username:
        smtp.login(username, password)
    smtp.sendmail(send_from, send_to, msg.as_string())
    smtp.close()


class SMTPBackend(Backend):
    def send(self, ev):
        for user in ev['users']:
            if not user.get('email'):
                continue

            subject = u'%s[P%s]%s' % (
                status2emoji(ev['status']),
                ev['level'],
                ev['title'],
            )

            send_mail(
                self.conf['send_from'], user['email'],
                subject, ev['text'],
                server=self.conf['server'],
                ssl=self.conf.get('ssl', False),
                username=self.conf['username'],
                password=self.conf['password'],
            )


EXPORT = SMTPBackend
