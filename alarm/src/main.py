# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- stdlib --
from functools import partial
import copy
import datetime
import json
import logging
import time

# -- third party --
from gevent.lock import RLock
import gevent
import gevent.pool
import redis

# -- own --
from state import State
import backend


# -- code --
pool = gevent.pool.Pool(20)
log = logging.getLogger('main')


def cook_event(ev):
    ev.pop('version', '')
    required = {
        u'id', u'time', u'level', u'status', u'endpoint',
        u'metric', u'tags', u'note', u'expected', u'actual',
        u'groups',
    }

    missing = required - (required & set(ev.keys()))

    if missing:
        raise Exception('Missing field: %s', list(missing))

    ts = datetime.datetime. \
        fromtimestamp(ev['time']). \
        strftime("%Y-%m-%d %H:%M:%S").decode('utf-8')

    final = [
        (u"Time", ts),
        (u"Metric", "%s == %s" % (ev['metric'], ev['actual'])),
    ]

    final.extend(ev['tags'].items())

    ev['text'] = u'\n'.join([u"%s: %s" % (k, v) for k, v in final])

    desc = ev.get('description')
    if desc:
        ev['text'] += '\n' + desc

    ev['formatted_time'] = ts
    ev['title'] = u'%s: %s' % (ev['endpoint'], ev['note'])

    teams = ev['groups']
    mapping = {t: State.teams.get(t, []) for t in teams}

    related_teams = {}
    for t, uids in mapping.items():
        for uid in uids:
            related_teams.setdefault(uid, set()).add(t)

    users = [State.users.get(u) for l in mapping.values() for u in l]
    users = [copy.deepcopy(u) for u in users if not u or u.get('threshold', 1000) > ev['level']]

    for u in users:
        u['related_groups'] = list(related_teams[u['id']])

    ev['users'] = users

    return ev


def get_relevant_backends(ev):
    lvl = str(ev['level'])
    return [
        State.backends[conf['backend']]
        for conf in State.strategies.values()
        if lvl in str(conf['level'])
    ]


class AlarmDFA(object):
    trans = set()
    ticking_states = set()

    def __init__(self):
        self.alarms = {}
        self.lock = RLock()

    def transition(when, action, valid, trans=trans, ticking_states=ticking_states):
        def decorate(f):
            trans.add(((when, action), (f, tuple(valid))))
            if action == 'TICK':
                ticking_states.add(when)

            return f

        return decorate

    def transit(self, id=None, action=None, ev=None):
        assert bool(id and action) ^ bool(ev)

        with self.lock:
            if ev:
                id = ev['id']
                event = ev['status']
                last = self.alarms.get(id)
                when = last['status'] if last else 'OK'
                action = ev['status']
            else:
                ev = self.alarms.get(id)
                if not ev:
                    log.info('AlarmDFA: Id %s does not exist', id)
                    return 'ERROR'
                when = ev['status']

            find = (when, action)
            for find in [(when, action), ('@ALWAYS', action)]:
                for cond, (handler, valid) in self.trans:
                    if find == cond:
                        log.debug('Handling %s by %s', find, handler.__name__)
                        new = handler(self, when, action, ev)
                        assert new in valid or new == 'ERROR', 'Invalid transition: %s handled by %s, want to go %s' % (find, handler.__name__, new)
                        ev['status'] = new

                        if new == 'OK':
                            self.alarms.pop(id, '')
                        else:
                            self.alarms[id] = ev

                        return new
            else:
                log.error("Can't handle %s in state %s", event, when)
                return 'ERROR'

    def tick(self):
        with self.lock:
            for k, v in list(self.alarms.iteritems()):
                if v['status'] in self.ticking_states:
                    self.transit(id=k, action='TICK')

    def loads(self, s):
        with self.lock:
            self.alarms = json.loads(s)

    def dumps(self):
        with self.lock:
            return json.dumps(self.alarms)

    def values(self):
        with self.lock:
            return self.alarms.values()

    @classmethod
    def _dump_dfa(cls):
        '$ dot -Tpng /tmp/dfa.dot > dfa.png'

        with open('/tmp/dfa.dot', 'w') as file:
            W = lambda s: file.write(s + '\n')
            W('digraph AlarmDFA {')
            for ((when, action), (f, valid)) in cls.trans:
                for v in valid:
                    # W('    "%s" -> "%s" [label="%s[%s]"]' % (when, v, action, f.__name__))
                    W('    "%s" -> "%s" [label="%s"]' % (when, v, action))
            W('}')

    # ---------
    @transition(when='OK',      action='PROBLEM', valid=['PROBLEM'])
    @transition(when='PROBLEM', action='PROBLEM', valid=['PROBLEM'])
    def do_send_alarm(self, when, action, ev):
        send_alarm(ev)
        return 'PROBLEM'

    @transition(when='OK', action='EVENT', valid=['OK'])
    def do_send_event(self, when, action, ev):
        send_alarm(ev)
        return 'OK'

    @transition(when='PROBLEM', action='OK', valid=['OK'])
    def do_recover(self, when, action, ev):
        send_alarm(ev)
        return 'OK'

    @transition(when='OK',       action='OK',      valid=['OK'])
    @transition(when='ACK',      action='PROBLEM', valid=['ACK'])
    @transition(when='FLAPPING', action='PROBLEM', valid=['FLAPPING'])
    def do_nothing(self, when, action, ev):
        return when

    @transition(when='ACK', action='OK', valid=['TIMEWAIT'])
    def do_recover_in_ack(self, when, action, ev):
        ev['timewait'] = time.time()
        send_alarm(ev)
        return 'TIMEWAIT'

    @transition(when='FLAPPING', action='OK', valid=['TIMEWAIT'])
    def do_recover_in_flapping(self, when, action, ev):
        ev['timewait'] = time.time()
        return 'TIMEWAIT'

    @transition(when='TIMEWAIT', action='PROBLEM', valid=['FLAPPING'])
    def do_timewait_resurrect(self, when, action, ev):
        ev.pop('timewait', '')
        return 'FLAPPING'

    @transition(when='TIMEWAIT', action='TICK', valid=['TIMEWAIT', 'OK'])
    def do_tw_recycle(self, when, action, ev):
        tw_sec = State.userconf.get('timewait_seconds', 0)
        if time.time() - ev['timewait'] < tw_sec:
            return 'TIMEWAIT'
        else:
            return 'OK'

    @transition(when='ACK',      action='TOGGLE_ACK', valid=['PROBLEM'])
    @transition(when='FLAPPING', action='TOGGLE_ACK', valid=['PROBLEM'])
    def do_to_problem(self, when, action, ev):
        return 'PROBLEM'

    @transition(when='PROBLEM', action='TOGGLE_ACK', valid=['ACK'])
    def do_to_ack(self, when, action, ev):
        return 'ACK'

    @transition(when='@ALWAYS', action='RESOLVE', valid=['OK'])
    def do_resolve(self, when, action, ev):
        return 'OK'

    # ---------

    del transition


def send_alarm(ev):
    backends = get_relevant_backends(ev)

    #log.info('alarm event %s' % ev)
    for p in backends:
        try:
            gevent.spawn(p.send, ev['users'], copy.deepcopy(ev))
        except Exception:
            log.exception('Error batch processing event')


def process_single_event(raw):
    ev = cook_event(raw)

    if not ev:
        log.warning("Can't handle event: %s", ev)
        return

    State.alarms.transit(ev=ev)


def alarm_tick():
    while True:
        gevent.sleep(1)
        State.alarms.tick()


def process_events():
    r = redis.from_url(State.config['redis'])
    queues = [
        'satori-events:%s' % i for i in range(15)
    ]

    while True:
        o, raw = r.blpop(queues)
        pool.spawn(process_single_event, json.loads(raw))
