# -*- coding: utf-8 -*-

# -- stdlib --
import gevent
import logging
import os
import subprocess
import time

# -- third party --
import yaml

# -- own --
from state import State
import backend
import hook


# -- code --
log = logging.getLogger('config')


def read_config(path):
    teams = {}
    users = {}
    strategies = {}
    others = {}

    base = os.path.join(path, 'alarm')
    for fn in os.listdir(base):
        if not fn.endswith('.yml') and not fn.endswith('.yaml'):
            continue

        with open(os.path.join(base, fn)) as f:
            conf = yaml.safe_load(f.read())

        if 'teams' in conf:
            assert not set(teams) & set(conf['teams'])
            teams.update(conf['teams'])
            conf.pop('teams', 0)
        # 'groups' is acronym of 'teams'
        if 'groups' in conf:
            assert not set(teams) & set(conf['groups'])
            teams.update(conf['groups'])
            conf.pop('groups', 0)
        if 'users' in conf:
            assert not set(users) & set(conf['users'])
            users.update(conf['users'])
            conf.pop('users', 0)
        if 'strategies' in conf:
            assert not set(strategies) & set(conf['strategies'])
            strategies.update(conf['strategies'])
            conf.pop('strategies', 0)

        others.update(conf)

    for k, v in users.items():
        v['id'] = k

    State.teams = teams
    State.users = users
    State.strategies = strategies
    State.userconf = others


def refresh():
    try:
        [bk.shutdown() for bk in State.backends.values()]
    except AttributeError:
        pass

    path = State.config['rules']

    read_config(path)
    backend.scrape_backends(path)
    hook.scrape_hooks(path)

    backends = {}
    for strategy in State.strategies.values():
        backends[strategy['backend']] = backend.from_string(
            strategy['backend']
        )(strategy)
    State.backends = backends


def watch_loop():
    rev = 'fresh'

    log.info('Users config watch loop started')

    while True:
        time.sleep(3)
        proc = subprocess.Popen(
            ['/usr/bin/git', 'rev-parse', 'HEAD'],
            cwd=State.config['rules'], stdout=subprocess.PIPE
        )
        code = proc.wait()
        if code != 0:
            log.error('Failed to git rev-parse, sleep for a while')
            time.sleep(60)
            continue

        new_rev = proc.stdout.read().strip()
        if rev == new_rev:
            continue

        refresh()

        log.info('Refreshing users conf: %s -> %s', rev, new_rev)
        rev = new_rev


def start_watch():
    gevent.spawn(watch_loop)
