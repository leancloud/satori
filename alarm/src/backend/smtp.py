# -*- coding: utf-8 -*-

# -- stdlib --
from email.MIMEMultipart import MIMEMultipart
from email.MIMEText import MIMEText
from email.Utils import formatdate
import smtplib

# -- third party --
# -- own --
from backend.common import register_backend

# -- code --


def send_mail(send_from, send_to, subject, text, files=[], server="localhost", username=None, password=None):
    msg = MIMEMultipart('alternative')
    msg.set_charset('utf-8')
    msg['From'] = send_from
    msg['To'] = send_to
    msg['Date'] = formatdate(localtime=True)
    msg['Subject'] = subject
    part = MIMEText(text)
    part.set_charset('utf-8')
    msg.attach(part)
    smtp = smtplib.SMTP(server)
    if username:
        smtp.login(username, password)
    smtp.sendmail(send_from, send_to, msg.as_string())
    smtp.close()


@register_backend
def smtp(conf, user, event):
    if not user['email']:
        return

    subject = u'%s[P%s]%s' % (
        u'😱' if event['status'] in ('PROBLEM', 'EVENT') else u'😅',
        event['level'],
        event['title'],
    )

    send_mail(
        conf['send_from'], user['email'],
        subject, event['text'],
        server=conf['server'],
        username=conf['username'],
        password=conf['password'],
    )
