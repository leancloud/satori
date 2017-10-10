.. _proc-java-heap:

Java 进程 OldGen 占用
=====================

收集指定 Java 进程的 OldGen （老代堆）占用

插件文件地址
    _metric/proc.java.cpu

插件类型
    接受参数，重复执行


.. warning::
   如果你的 Java 进程是在容器内运行的，那么你必须在容器内进行收集。
   插件是通过 ``jstat -gccause <pid>`` 命令收集的，这个命令需要读取 :file:`/tmp/hsperfdata_*` 文件中的信息。


插件参数
--------

+---------+------------------------------------+
| 参数    | 功能                               |
+=========+====================================+
| cmdline | 被监控进程的命令行正则表达式 [#]_  |
+---------+------------------------------------+
| name    | 这个监控的名字，用来区分其他的监控 |
+---------+------------------------------------+

.. [#] 命令行是从 /proc/ ``pid`` /cmdline 读取并且将其中的 ``\x00`` 替换成空格后进行匹配。
       匹配是任意位置的，需要限定从一开始匹配的话请在正则前面加上 ``^``


上报的监控值
------------

proc.java.heap
   :意义: 指定的 Java 进程的 OldGen 占用百分比
   :取值: 0 - 100
   :Tags: {"pid": " ``指定进程的 pid`` ", "name": " ``指定的监控名字`` "}

监控规则样例
------------

.. code-block:: clojure

  (def infra-es-rules
    (sdo
      (where (host #"^es\d$")
        (plugin-metric "proc.java.heap" 30
          {:name "elasticsearch", :cmdline "org.elasticsearch.bootstrap.Elasticsearch"})

        (where (and (service "proc.java.heap")
                    (= (:name event) "elasticsearch"))
          (by [:host :region]
            (set-state-gapped (> 99.8) (< 95)
              (runs 3 :state
                (should-alarm-every 120
                  (! {:note "ElasticSearch OldGen 满了！"
                      :level 1
                      :expected true
                      :outstanding-tags [:host :name]
                      :groups [:operation :api]})))))))))
