#!/usr/bin/python

import time

l = [0]
for i in range(1000):
    l = l * 2
    time.sleep(0.1)
