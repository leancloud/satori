.. _cross:

集群交叉检测
************

Ping 检测
=========

在指定的机器上 Ping 整个集群的机器

插件文件地址
    cross.ping

插件类型
    接受参数，重复执行

插件参数
--------

+--------+--------------------+
| 参数   | 功能               |
+========+====================+
| region | 希望检测的集群名称 |
+--------+--------------------+

上报的监控值
------------

cross.ping.alive
    :意义: 机器是否可以 Ping 通
    :取值: 0 代表失败，1 代表成功
    :Tags: {"from": "``插件执行的机器的 hostname``"}，另外 ``endpoint`` (对应 riemann 中的 ``:host``) 值是被检测的机器的 hostname

cross.ping.latency
    :意义: 机器的 Ping 延迟
    :取值: 0-无上限，单位 ms
    :Tags: {"from": "``插件执行的机器的 hostname``"}，另外 ``endpoint`` (对应 riemann 中的 ``:host``) 值是被检测的机器的 hostname


.. note::
    插件需要在执行插件的机器上安装 ``fping`` 工具

.. warning::
   插件会通过 :file:`utils/region.py` 中的 ``nodes_of`` 函数取得指定集群的机器列表，
   你需要自己实现这个函数


监控规则样例
------------

.. code-block:: clojure

    (where (host "hosts" "which" "perform" "ping")
      (plugin "cross.ping" 15 {:region "office"})

    (where (service "cross.ping.alive")
      (by :host
        (judge (< 1)
          (runs 3 :state
            (alarm-every 2 :min
              (! {:note "Ping 不通了！"
                  :level 1
                  :expected 1
                  :outstanding-tags [:region]
                  :groups [:operation]}))))))

Agent 存活检测
==============

在指定的机器上 Satori agent 是否存活

插件文件地址
    cross.agent

插件类型
    接受参数，重复执行

插件参数
--------

+--------+--------------------+
| 参数   | 功能               |
+========+====================+
| region | 希望检测的集群名称 |
+--------+--------------------+

上报的监控值
------------

cross.agent.alive
    :意义: 指定机器上的 agent 是否存活
    :取值: 0 代表不存活，1 代表存活
    :Tags: {"from": "``插件执行的机器的 hostname``"}，另外 ``endpoint`` (对应 riemann 中的 ``:host``) 值是被检测的机器的 hostname

.. warning::
   插件会通过 :file:`utils/region.py` 中的 ``nodes_of`` 函数取得指定集群的机器列表，
   你需要自己实现这个函数


监控规则样例
------------

.. code-block:: clojure

    (where (host "hosts" "which" "perform" "ping")
      (plugin "cross.agent" 15 {:region "office"})

    (where (service "cross.agent.alive")
      (by :host
        (judge (< 1)
          (runs 3 :state
            (alarm-every 2 :min
              (! {:note "Satori Agent 不响应了！"
                  :level 1
                  :expected 1
                  :outstanding-tags [:region]
                  :groups [:operation]}))))))
