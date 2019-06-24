.. _writing-rules:

编写规则简介
============

规则文件的组织
--------------

进行报警判定的规则是由 Clojure 语言编写的，放在 :file:`rules` 文件夹中。

这些规则将由 Riemann_ 进行执行， Riemann_ 会接受从 transfer 发来的指标，交给编写的规则进行判定，然后再发送给 alarm 进行报警。

规则的 ``.clj`` 文件需按照约定放在单层的文件夹下，并且 namespace 需要与文件夹和文件名对应。
文件中所有以 ``-rules`` 结尾的公开变量都会被当做 riemann 流来做判定。

.. _Riemann: http://riemann.io


Copy & Paste 的正确姿势
-----------------------

因为通常规则文件长得都差不多，所以推荐找一个附带的规则 Copy & Paste 一下，
比如，我想添加一个 ``foo/bar.clj`` 作为 foo 服务下关于 bar 的监控，
那么就随便找一个文件夹中的规则（最外层的不行！），然后修改 ns 为 ``foo.bar`` ，
然后挑一个规则复制进来，规则的名字改成 ``foo-bar-rules`` ，然后按照需求修改规则就好了。

.. code-block:: clojure

    (ns infra.memcache ; 这里要修改
      (:use riemann.streams
            agent-plugin
            alarm))

    (def infra-memcache-rules ; 这里名字要修改
      (where (host #"cache\d+$") ; 规则当然也要改
        (plugin-dir "memcache")
        (plugin "net.port.listen" 30 {:port 11211})

        (where (and (service "net.port.listen")
                    (= (:port event) 11211))
          (by :host
            (judge-gapped (< 1) (> 0)
              (alarm-every 2 :min
                (! {:note "memcache 端口不监听了！"
                    :level 0
                    :groups [:operation :api]})))))))

修改成

.. code-block:: clojure

    (ns foo.bar
      (:use riemann.streams
            agent-plugin
            alarm))

    (def foo-bar-rules ; 随便叫什么但是一定要以 -rules 结尾
      (where (host "host-which-run-foo-bar")
        (plugin-dir "foo/bar") ; 要运行 foo/bar 目录里的插件
        (plugin "net.port.listen" 30 {:port 12345}) ; 运行 net.port.listen 插件，30秒一次，以及参数

        (where (and (service "net.port.listen")
                    (= (:port event) 12345)) ; 匹配 net.port.listen 插件收集的 metric
          (by :host ; 按照 host 分成子流（这个例子里就只有一个 host 所以并不会分）
            (judge-gapped (< 1) (> 0)  ; 根据条件设置事件的 :state
              (alarm-every 2 :min ; 如果是 :problem 每2分钟发一次报警。
                (! {:note "foobar 服务端口不监听了！"
                    :level 0
                    :groups [:operation :api]})))))))

commit 然后 push 上去就会生效了。


事件流的组织
------------

所有的事件汇集到 Riemann 之后，你看到的是一个由所有机器、服务收集上来的指标组成的一个单一的事件流，
或者说，每一个 ``(def xxx-yyy-rules ...)`` 都能够看到整个集群（或者n个集群，看你是怎么部署的了）所有的指标，
你需要做的是通过过滤、变换等等手段最后将你关心的东西过滤出来，判定问题并告诉 alarm。


错误处理
--------

可以在 :file:`riemann.config` 中找到相关的规则，可以修改报警级别和报警组（与正常的规则一样）。

监控规则 Reload 失败后还会用老规则，所以不用担心。

发送报警
--------

流向 :ref:`exclaimation-mark` 流的事件都会发送到 alarm 进行报警。
但是不能直接把事件喂给 :ref:`exclaimation-mark` ，因为
:ref:`exclaimation-mark` 并不知道这个事件是正常还是有问题的状态， 所以需要指定事件的状态是
``:ok`` 还是 ``:problem`` 。 通常可以用 :ref:`judge` 和
:ref:`judge-gapped` 这两个流来完成。

Riemann 提供的文档
------------------

在 Riemann 中可用的流不仅仅是这里介绍的，还可以参考 `Riemann 官方文档`_ ，还有很多不常用函数/流在里面有介绍。

.. _`Riemann 官方文档`: http://riemann.io/api/index.html

.. meh
    最简单的一个例子：
    .. code-block:: clojure
        (def app-important-rules
        (where (service "service.very.important.latency")
            (judge-gapped (> 1000) (< 100)  ; 当 latency 超过 1000 后报警，回落到 100 以下变成正常状态
                (! {:note "报警标题，标题对于一个特定的报警是不能变的（不要把报警的数据编码在这里面）"
                    :level 1  ;报警级别, 0最高，6最小。报警级别影响报警方式。
                    :event? false  ; 可选，是不是事件（而不是状态）。默认 false。如果是事件的话，只会发报警，不会记录状态（alarm插件里看不到）。
                    :expected 233  ; 可选，期望值，暂时没用到
                    :outstanding-tags [:region :mount]  ; 可选，相关的tag，写在这里的 tag 会用于区分不同的事件，以及显示在报警内容中, 不填的话默认是所有的tag
                    :groups [:operation]})  ; groups 是在规则仓库的 alarm 配置里管理的)
    报警级别是在 alarm 的配置中定义的，具体可以看一下 ``02-config-alarm.md``
    文件。
