# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- prioritized --
from gevent import monkey
monkey.patch_all()

# -- stdlib --
import argparse
import datetime
import logging
import sys

# -- third party --
from raven.handlers.logging import SentryHandler
from raven.transport.gevent import GeventedHTTPTransport
import raven
import yaml

# -- own --
from state import State
from utils import spawn_autorestart
import config
import main
import save


# -- code --
def patch_gevent_hub_print_exception():
    from gevent.hub import Hub

    def print_exception(self, context, type, value, tb):
        import logging
        log = logging.getLogger('exception')
        log.error(
            '%s failed with %s',
            context, getattr(type, '__name__', 'exception'),
            exc_info=(type, value, tb),
        )

    Hub.print_exception = print_exception


class ServerLogFormatter(logging.Formatter):
    def format(self, rec):

        if rec.exc_info:
            s = []
            s.append('>>>>>>' + '-' * 74)
            s.append(self._format(rec))
            import traceback
            s.append(u''.join(traceback.format_exception(*rec.exc_info)).strip())
            s.append('<<<<<<' + '-' * 74)
            return u'\n'.join(s)
        else:
            return self._format(rec)

    def _format(self, rec):
        import time
        return u'[%s %s] %s' % (
            rec.levelname[0],
            time.strftime('%y%m%d %H:%M:%S'),
            (rec.msg % rec.args) if isinstance(rec.msg, basestring) else repr((rec.msg, rec.args))
        )


def init_logging(level='INFO'):
    root = logging.getLogger()
    root.setLevel(0)

    patch_gevent_hub_print_exception()
    if State.config.get('sentry'):
        hdlr = SentryHandler(raven.Client(State.config['sentry'], transport=GeventedHTTPTransport))
        hdlr.setLevel(logging.ERROR)
        root.addHandler(hdlr)

    fmter = ServerLogFormatter()

    hdlr = logging.StreamHandler(sys.stdout)
    hdlr.setLevel(getattr(logging, level))
    hdlr.setFormatter(fmter)

    root.addHandler(hdlr)

    root.info(datetime.datetime.now().strftime("%Y-%m-%d %H:%M"))
    root.info('==============================================')


def start():
    parser = argparse.ArgumentParser('falcon-alarm')
    parser.add_argument('--config', help='Config file')
    parser.add_argument('--log', type=str, default="INFO", help='Config file')
    options = parser.parse_args()

    State.options = options
    State.config = yaml.load(open(options.config).read())

    init_logging(options.log)

    config.refresh()
    config.start_watch()

    State.alarms = main.AlarmDFA()

    save.load()
    save.start_periodically_dump()

    spawn_autorestart(main.process_events)
    spawn_autorestart(main.alarm_tick)

    import restapi
    restapi.serve()


if __name__ == '__main__':
    start()
