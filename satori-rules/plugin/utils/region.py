# -*- coding: utf-8 -*-
# from __future__ import absolute_import, division, print_function, unicode_literals

# -- stdlib --
# -- third party --
try:
    import ldap
except ImportError:
    raise ImportError("No module named ldap, please install via `apt install python-ldap`")


# -- own --


# -- code --
def nodes_of(region):
    svr = ldap.initialize('ldap://ldap.example.com')
    svr.simple_bind_s('uid=user,cn=users,cn=accounts,dc=in,dc=example,dc=com', 'example-password')

    return [i[1]['fqdn'][0] for i in svr.search_s(
        'cn=computers,cn=accounts,dc=in,dc=example,dc=com',
        ldap.SCOPE_ONELEVEL,
        'cn=*.%s.in.example.com' % region,
        ['fqdn']
    )]
