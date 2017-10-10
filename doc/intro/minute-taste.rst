.. _minute-taste:

一分钟快速感受
--------------

.. _common-rule-demo:

常见需求
~~~~~~~~

.. code-block:: clojure

    (def infra-mongodb-rules
      ; 在 mongo1 mongo2 mongo3 ... 上做监控
      (where (host #"^mongo\d+")

        ; 执行 mongodb 目录里的插件（里面有收集 mongodb 指标的脚本）
        (plugin-dir "mongodb")

        ; 每 30s 收集 27018 端口是不是在监听
        (plugin-metric "net.port.listen" 30 {:port 27018})

        ; 过滤一下 mongodb 可用连接数的 metric（上面插件收集的）
        (where (service "mongodb.connections.available")
          ; 按照 host（插件中是 endpoint）分离监控项，并分别判定
          (by :host
            ; 报警在监控值 < 10000 时触发，在 > 50000 时恢复
            (set-state-gapped (< 10000) (> 50000)
              ; 600 秒内只报告一次
              (should-alarm-every 600
                ; 报告的标题、级别（影响报警方式）、报告给 operation 组和 api 组
                (! {:note "mongod 可用连接数 < 10000 ！"
                    :level 1
                    :groups [:operation :api]})))))

      ; 另一个监控项
      (where (service "mongodb.globalLock.currentQueue.total")
        (by :host
          (set-state-gapped (> 250) (< 50)
            (should-alarm-every 600
              (! {:note "MongoDB 队列长度 > 250"
                  :level 1
                  :groups [:operation :api]})))))))


.. code-block:: bash

    cd /path/to/rules/repo  # 规则是放在 git 仓库中的
    git commit -a -m 'Add mongodb rules'
    git push  # 然后就生效了

.. note::

    完整版请看 :file:`satori-rules/rules/infra/mongodb.clj`

    这个模板可以满足常见的需求，真正编写规则的时候可以复制粘贴这一份规则并做微调



.. _complex-rule-demo:

复杂需求
~~~~~~~~

这是一个监控队列堆积的规则。
队列做过 sharding，分布在多个机器上。
但是有好几个数据中心，要分别报告每个数据中心队列情况。
堆积的定义是：在一定时间内，队列非空，而且队列元素数量没有下降。

.. code-block:: clojure

    (def infra-kestrel-rules
      ; 在所有的队列机器上做监控
      (where (host #"kestrel\d+$")
        ; 执行队列相关的监控脚本（插件）
        (plugin-dir "kestrel")

        ; 过滤『队列项目数量』的 metric
        (where (service "kestrel_queue.items")
          ; 按照队列名和数据中心分离监控项，并分别判定
          (by [:queue :region]
            ; 将传递下来的监控项暂存 60 秒，然后打包（变成监控项的数组）再向下传递
            (fixed-time-window 60
              ; 将打包传递下来的监控项做聚集：将所有的 metric 值都加起来。
              ; 因为队列监控的插件是每 60 秒报告一次，并且之前已经按照队列名和数据中心分成子流了，
              ; 所以这里将所有 metric 都加起来以后，获得的是单个数据中心单个队列的项目数量。
              ; 聚集后会由监控项数组变成单个的监控项。
              (aggregate +
                ; 将传递下来的聚集后的监控项放到滑动窗口里，打包向下传递。
                ; 这样传递下去的，就是一个过去 600 秒单个数据中心单个队列的项目数量的监控项数组。
                (moving-event-window 10
                  ; 如果已经集满了 10 个，而且这 10 个监控项中都不为 0 （队列非空）
                  (where (and (>= (count event) 10)
                              (every? pos? (map :metric event)))
                    ; 再次做聚集：比较一下是不是全部 10 个数量都是先来的小于等于后来的（是不是堆积）
                    (aggregate <=
                      ; 如果结果是 true，那么认为现在是有问题的
                      (set-state (= true)
                        ; 每 1800 秒告警一次
                        (should-alarm-every 1800
                          ; 这里 outstanding-tags 是用来区分报警的，
                          ; 即如果 region 的值不一样，那么就会被当做不同的报警
                          (! {:note #(str "队列 " (:queue %) " 正在堆积！")
                              :level 2
                              :outstanding-tags [:region]
                              :groups [:operation]}))))))))))))

.. note::

    这是一个简化了的版本，完整版可以看 :file:`satori-rules/rules/infra/kestrel.clj`

