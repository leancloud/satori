#!/usr/bin/python
# -*- coding: utf-8 -*-

from setuptools import setup, find_packages

install_requires = [
    'bottle',
    'gevent',
    'redis',
    'pyaml',
    'requests',
    'raven',
]

entry_points = {
    'console_scripts': [
        'start = entry:start',
        'inject = tools.inject:main',
    ]
}

setup(
    name="falcon-alarm",
    version="1.0.0",
    url='https://leancloud.cn',
    license='Propritary',
    description='Satori alarm module',
    author='proton',
    packages=find_packages('src'),
    package_dir={'': 'src'},
    install_requires=install_requires,
    entry_points=entry_points,
)
