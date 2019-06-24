.. _ping:

Ping 检测
=========

在指定的机器上 Ping 整个集群的机器

插件文件地址
    ping/*

插件类型
    不接受参数（直接执行），重复执行


上报的监控值
------------

agent.ping
    :意义: 访问其他机器的 satori-agent 的 ``/v1/ping`` 接口的结果
    :提供: :file:`ping/60_agent_ping.py`
    :取值: 0 代表失败，1 代表成功
    :Tags: {"from": "``插件执行的机器的 hostname``"}，另外 ``endpoint`` (对应 riemann 中的 ``:host``) 值是被检测的机器的 hostname

ping.loss
    :意义: ICMP Ping 其他机器的丢包率
    :提供: :file:`ping/60_fping.rb`
    :取值: 0 - 100.0，浮点数
    :Tags: {"from": "``插件执行的机器的 hostname``"}，另外 ``endpoint`` (对应 riemann 中的 ``:host``) 值是被检测的机器的 hostname

ping.latency
    :意义: ICMP Ping 其他机器的延迟
    :提供: :file:`ping/60_fping.rb`
    :取值: 0 - 无上限，浮点数，单位：ms
    :Tags: {"from": "``插件执行的机器的 hostname``"}，另外 ``endpoint`` (对应 riemann 中的 ``:host``) 值是被检测的机器的 hostname

ping.alive
    :意义: ICMP Ping 其他机器是否存活
    :提供: :file:`ping/60_fping.rb`
    :取值: 0 - 无上限，浮点数，单位：ms
    :Tags: 无


.. note::
    ``ping/60_fping.rb`` 插件需要在执行插件的机器上安装 ``fping`` 工具

.. warning::
    因为 LeanCloud 内部在用 Puppet，所以机器信息是从 PuppetDB 拉取的。如果决定使用这个插件，你需要修改插件来对接自己的 CMDB 拉取机器信息！


监控规则样例
------------

.. code-block:: clojure

    (where (host "hosts" "which" "perform" "ping")
      (plugin-dir "ping"))

    (where (service "ping.alive")
      (by :host
        (judge (< 1)
          (runs 3 :state
            (should-alarm-every 120
              (! {:note "Ping 不通了！"
                  :level 1
                  :expected 1
                  :outstanding-tags [:region]
                  :groups host->group}))))))
