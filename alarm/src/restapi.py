# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- stdlib --
# -- third party --
from bottle import route, run

# -- own --
from state import State


# -- code --
@route('/alarms')
def alarms():
    return {'alarms': State.alarms.values()}


@route('/alarms/<ids>', method='DELETE')
def resolve(ids):
    ids = ids.split(',')
    dfa = State.alarms
    for i in ids:
        dfa.transit(id=i, action='RESOLVE')

    return {}


@route('/alarms/<id>/toggle-ack', method='POST')
def toggle_ack(id):
    new = State.alarms.transit(id=id, action='TOGGLE_ACK')
    return {'new-state': new}


@route('/reload', method='POST')
def reload():
    from entry import read_users
    State.teams, State.users = read_users(State.config['rules'])
    return True


def serve():
    run(
        server='gevent',
        host=State.config['host'],
        port=State.config['port'],
        debug=True,
    )
