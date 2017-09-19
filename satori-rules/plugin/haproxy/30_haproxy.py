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

# -- own --

# -- code --
ts = int(time.time())

FIELDS = [
    # "pxname",      # [LFBS]: proxy name
    # "svname",      # [LFBS]: service name (FRONTEND for frontend, BACKEND for backend, any name for server/listener)
    "qcur",          # [..BS]: current queued requests. For the backend this reports the number queued without a server assigned.
    "qmax",          # [..BS]: max value of qcur
    "scur",          # [LFBS]: current sessions
    "smax",          # [LFBS]: max sessions
    "slim",          # [LFBS]: configured session limit
    "stot",          # [LFBS]: cumulative number of connections
    "bin",           # [LFBS]: bytes in
    "bout",          # [LFBS]: bytes out
    "dreq",          # [LFB.]: requests denied because of security concerns.
    "dresp",         # [LFBS]: responses denied because of security concerns.
    "ereq",          # [LF..]: request errors.
    "econ",          # [..BS]: number of requests that encountered an error trying to connect to a backend server.
    "eresp",         # [..BS]: response errors. srv_abrt will be counted here also.
    "wretr",         # [..BS]: number of times a connection to a server was retried.
    "wredis",        # [..BS]: number of times a request was redispatched to another server.
    "status",        # [LFBS]: status (UP/DOWN/NOLB/MAINT/MAINT(via)...)
    "weight",        # [..BS]: total weight (backend), server weight (server)
    "act",           # [..BS]: number of active servers (backend), server is active (server)
    "bck",           # [..BS]: number of backup servers (backend), server is backup (server)
    "chkfail",       # [...S]: number of failed checks. (Only counts checks failed when the server is up.)
    "chkdown",       # [..BS]: number of UP->DOWN transitions.
    "lastchg",       # [..BS]: number of seconds since the last UP<->DOWN transition
    "downtime",      # [..BS]: total downtime (in seconds).
    "qlimit",        # [...S]: configured maxqueue for the server, or nothing in the value is 0 (default, meaning no limit)
    "-",             # [LFBS]: process id (0 for first instance, 1 for second, ...)
    "-",             # [LFBS]: unique proxy id
    "-",             # [L..S]: server id (unique inside a proxy)
    "throttle",      # [...S]: current throttle percentage for the server, when slowstart is active, or no value if not in slowstart.
    "-",             # [..BS]: total number of times a server was selected, either for new sessions, or when re-dispatching.
    "-",             # [...S]: id of proxy/server if tracking is enabled.
    "-",             # [LFBS]: (0=frontend, 1=backend, 2=server, 3=socket/listener)
    "rate",          # [.FBS]: number of sessions per second over last elapsed second
    "rate_lim",      # [.F..]: configured limit on new sessions per second
    "rate_max",      # [.FBS]: max number of new sessions per second
    "check_status",  # [...S]: status of last health check, one of:
    "check_code",    # [...S]: layer5-7 code, if available
    "-",             # [...S]: time in ms took to finish last health check
    "hrsp_1xx",      # [.FBS]: http responses with 1xx code
    "hrsp_2xx",      # [.FBS]: http responses with 2xx code
    "hrsp_3xx",      # [.FBS]: http responses with 3xx code
    "hrsp_4xx",      # [.FBS]: http responses with 4xx code
    "hrsp_5xx",      # [.FBS]: http responses with 5xx code
    "hrsp_other",    # [.FBS]: http responses with other codes (protocol error)
    "hanafail",      # [...S]: failed health checks details
    "req_rate",      # [.F..]: HTTP requests per second over last elapsed second
    "req_rate_max",  # [.F..]: max number of HTTP requests per second observed
    "req_tot",       # [.F..]: total number of HTTP requests received
    "cli_abrt",      # [..BS]: number of data transfers aborted by the client
    "srv_abrt",      # [..BS]: number of data transfers aborted by the server (inc. in eresp)
    "comp_in",       # [.FB.]: number of HTTP response bytes fed to the compressor
    "comp_out",      # [.FB.]: number of HTTP response bytes emitted by the compressor
    "comp_byp",      # [.FB.]: number of bytes that bypassed the HTTP compressor (CPU/BW limit)
    "comp_rsp",      # [.FB.]: number of HTTP responses that were compressed
    "lastsess",      # [..BS]: number of seconds since last session assigned to server/backend
    "last_chk",      # [...S]: last health check contents or textual error
    "last_agt",      # [...S]: last agent check contents or textual error
    "qtime",         # [..BS]: the average queue time in ms over the 1024 last requests
    "ctime",         # [..BS]: the average connect time in ms over the 1024 last requests
    "rtime",         # [..BS]: the average response time in ms over the 1024 last requests (0 for TCP)
    "ttime",         # [..BS]: the average total session time in ms over the 1024 last requests
]

rst = []

try:
    resp = ''
    s = socket.socket(socket.AF_UNIX)
    s.connect('/run/haproxy/admin.sock')
    s.sendall('show stat\n')
    while True:
        frag = s.recv(10000)
        if not frag:
            break
        resp += frag

except Exception:
    pass

for entry in resp.strip().split('\n'):
    fields = entry.split(',')
    px, sv = fields[:2]
    d = {}
    for n, v in zip(FIELDS, fields[2:]):
        if n == '-' or v == '' or not v.isdigit():
            continue

        v = float(v)
        d[n] = v

        rst.append({
            "metric": "haproxy.%s" % n,
            "timestamp": ts,
            "value": v,
            "tags": {
                "proxy": px, "proxy-srv": sv,
            },
        })

    if 'slim' in d and sv == 'FRONTEND':
        rst.append({
            "metric": "haproxy.sratio",
            "timestamp": ts,
            "value": d['scur'] / d['slim'],
            "tags": {
                "proxy": px, "proxy-srv": sv,
            },
        })


print json.dumps(rst)
