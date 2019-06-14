# -*- coding: utf-8 -*-

# -- stdlib --
# -- third party --
from flask import Flask, jsonify

# -- own --
from state import State


# -- code --
app = Flask("alarm")


@app.route('/alarms')
def alarms():
    return jsonify({'alarms': list(State.alarms.values())})


@app.route('/alarms/<ids>', methods=['DELETE'])
def resolve(ids):
    ids = ids.split(',')
    dfa = State.alarms
    for i in ids:
        dfa.transit(id=i, action='RESOLVE')

    return jsonify({})


@app.route('/alarms/<id>/toggle-ack', methods=['POST'])
def toggle_ack(id):
    new = State.alarms.transit(id=id, action='TOGGLE_ACK')
    return jsonify({'new-state': new})


@app.route('/reload', methods=['POST'])
def reload():
    from entry import read_users
    State.teams, State.users = read_users(State.config['rules'])
    return jsonify({})


def serve():
    from gevent.pywsgi import WSGIServer
    svr = WSGIServer(f'{State.config["host"]}:{State.config["port"]}', app)
    svr.serve_forever()
