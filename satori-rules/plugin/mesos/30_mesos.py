#!/usr/bin/python
# -*- coding: utf-8 -*-
# -- prioritized --
import sys
import os.path
sys.path.append(os.path.join(os.path.dirname(__file__), '../libs'))

# -- stdlib --
import json
import socket
import time

# -- third party --
import requests

# -- own --

# -- code --
endpoint = socket.gethostname()
ts = int(time.time())

raw = {}

# master
try:
    ip = open('/etc/mesos-master/ip').read().strip()
    info = requests.get('http://%s:5050/metrics/snapshot' % ip, timeout=3).json()
    raw.update(info)
    raw['master/collect_success'] = 1
except Exception:
    raw['master/collect_success'] = 0

# slave
try:
    ip = open('/etc/mesos-slave/ip').read().strip()
    info = requests.get('http://%s:5051/metrics/snapshot' % ip, timeout=3).json()
    raw.update(info)
    raw['slave/collect_success'] = 1
except Exception:
    raw['slave/collect_success'] = 0


counters = {
    'containerizer/mesos/container_destroy_errors',
    'master/dropped_messages',
    'master/invalid_executor_to_framework_messages',
    'master/invalid_framework_to_executor_messages',
    'master/invalid_status_update_acknowledgements',
    'master/invalid_status_updates',
    'master/messages_authenticate',
    'master/messages_deactivate_framework',
    'master/messages_decline_offers',
    'master/messages_executor_to_framework',
    'master/messages_exited_executor',
    'master/messages_framework_to_executor',
    'master/messages_kill_task',
    'master/messages_launch_tasks',
    'master/messages_reconcile_tasks',
    'master/messages_register_framework',
    'master/messages_register_slave',
    'master/messages_reregister_framework',
    'master/messages_reregister_slave',
    'master/messages_resource_request',
    'master/messages_revive_offers',
    'master/messages_status_update',
    'master/messages_status_update_acknowledgement',
    'master/messages_unregister_framework',
    'master/messages_unregister_slave',
    'master/messages_update_slave',
    'master/recovery_slave_removals',
    'master/slave_registrations',
    'master/slave_removals',
    'master/slave_removals/reason_registered',
    'master/slave_removals/reason_unhealthy',
    'master/slave_removals/reason_unregistered',
    'master/slave_reregistrations',
    'master/slave_shutdowns_canceled',
    'master/slave_shutdowns_completed',
    'master/slave_shutdowns_scheduled',
    'master/task_lost/source_master/reason_invalid_offers',
    'master/task_lost/source_master/reason_slave_removed',
    'master/task_lost/source_slave/reason_executor_terminated',
    'master/tasks_error',
    'master/tasks_failed',
    'master/tasks_finished',
    'master/tasks_killed',
    'master/tasks_lost',
    'master/valid_executor_to_framework_messages',
    'master/valid_framework_to_executor_messages',
    'master/valid_status_update_acknowledgements',
    'master/valid_status_updates',
    'slave/container_launch_errors',
    'slave/executors_preempted',
    'slave/executors_terminated',
    'slave/invalid_framework_messages',
    'slave/invalid_status_updates',
    'slave/tasks_failed',
    'slave/tasks_finished',
    'slave/tasks_killed',
    'slave/tasks_lost',
    'slave/valid_framework_messages',
    'slave/valid_status_updates',
}

result = []
for k, v in raw.iteritems():
    result.append({
        "metric": "mesos.%s" % k.replace('/', '.'),
        "endpoint": endpoint,
        "timestamp": ts,
        "step": 30,
        "value": v,
    })

print json.dumps(result)
