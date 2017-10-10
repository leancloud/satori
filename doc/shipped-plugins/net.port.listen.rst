.. _net-port-listen:

端口监听
========

收集指定端口是否被监听

插件文件地址
    _metric/net.port.listen

插件类型
    接受参数，重复执行


插件参数
--------

+-----------+------------------------------------------------------------+
| 参数      | 功能                                                       |
+===========+============================================================+
| port      | 希望监听的端口                                             |
+-----------+------------------------------------------------------------+
| name      | **可选** ，这个监控的名字，用来区分其他的端口监听监控 [#]_ |
+-----------+------------------------------------------------------------+

.. [#] ``port`` 也会放到 tags 里，所以 ``name`` 做成了可选的。
       出于语义上的考虑，还是建议给一个名字，用名字来区分。


上报的监控值
------------

net.port.listen
   :意义: 指定的端口是否正在监听
   :取值: 0 和 1，分别表示没有监听和在监听
   :Tags: {"port": " ``指定的端口`` ", "name": " ``指定的监控名字`` "}

监控规则样例
------------

.. code-block:: clojure

    (def infra-redis-rules
      (sdo
        (where (host #"^redis-.+$")
          (plugin-metric "net.port.listen" 30 {:port 6379 :name "redis-port"})

          (where (and (service "net.port.listen")
                      (= (:name event) "redis-port"))
            (by [:host :port]
              (set-state (< 1)
                (runs 3 :state
                  (should-alarm-every 120
                    (! {:note "Redis 端口不监听了"
                        :level 1
                        :expected 3
                        :groups [:operation :api]})))))))))
