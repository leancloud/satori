.. _writing-plugin:

编写插件
========

插件分为两种，一种接受参数，另外一种不接受。

带参数的插件
------------

所有接受参数的插件都在 :file:`plugin` 目录里，命名需要跟报告的
metric 一致， 比如 :file:`plugin/net.port.listen` 会上报
``net.port.listen`` 这个 metric 。 接受参数的插件需要在在运行的时候通过
``stdin`` 读取 json 格式的参数：

.. code-block:: json

    [
      {"_metric": "net.port.listen", "_step": 30, "port": 6379},
      {"_metric": "net.port.listen", "_step": 30, "port": 3306}
    ]

.. warning::

    注意这里是 object 列表，不是单一的 object

参数是在规则配置中指定的，比如

.. code-block:: clojure

    (def mysql-and-redis-rules
      (where (host "mysql-and-redis-host")
        (plugin "net.port.listen" 30 {:port 3306})
        (plugin "net.port.listen" 30 {:port 6379})
        (your-other-rules ...)))

这里 :ref:`plugin` 的参数分别是插件名字、采集周期(step)、插件的参数。
采集周期如果是 ``0`` 则代表这个插件是持续运行的，agent 不会设置超时时间将插件杀死。
插件名字和采集周期会在参数里以 ``_metric`` 和 ``_step`` 作为 key 传递进去。

之后，插件需要输出如下格式的 json :

.. code-block:: json

    [
      {
        "endpoint": "mysql-and-redis-host",
        "metric": "net.port.listen",
        "value": 1,
        "timestamp": 1431349763,
        "tags": {"port": 3306},
        "step": 30
      },
      {
        "endpoint": "mysql-and-redis-host",
        "metric": "net.port.listen",
        "value": 0,
        "timestamp": 1431349763,
        "tags": {"port": 6379},
        "step": 30
      }
    ]

插件参数（比如上个例子中的 ``port`` ）可以任意添加。

.. warning::

    注意这里是 object 的 list，不是单一的 object。

    文档中为了清晰，json 是多行的，实际编写插件的时候 **同一个 list 请保持单行输出** ，
    因为 Satori 接受多行的输出（每一行是一个 list），需要通过换行来作为边界。

+-----------+------------------------------------------------------------+
| Key       | Value                                                      |
+===========+============================================================+
| endpoint  | **可选** [#]_ ，机器名，在 riemann 对应事件的 ``:host`` 。 |
+-----------+------------------------------------------------------------+
| tags      | **可选** [#]_ ，标签。在 riemann 中会直接附加到 event 上。 |
+-----------+------------------------------------------------------------+
| timestamp | 时间戳，事件发生的时间，以秒为单位的 UNIX 时间戳。         |
+-----------+------------------------------------------------------------+
| metric    | 监控项的名字。在 riemann 中对应事件的 ``:service`` 表示。  |
+-----------+------------------------------------------------------------+
| value     | 监控项的值。在 riemann 中对应事件的 ``:metric`` 。         |
+-----------+------------------------------------------------------------+
| step      | 采集周期，单位是秒。                                       |
+-----------+------------------------------------------------------------+

.. [#] 缺失则 agent 会按照自身配置填充这个值。 如果不需要控制这个值推荐省略让 agent 填充。
.. [#] 缺失则不会附加额外的 tag 。

.. note::

    插件的输出格式基本与 Open-Falcon 的插件格式相同，但是 tags 是个 object，
    不是拼接的字符串。

    插件中的 Key 跟 riemann 中对应不上是一个比较恼人的坑，需要特别注意。


无参数的插件
------------

无参数插件的命名需要类似于 ``30_nginx.py`` 这样的， ``_``
前面需要是数字，表示采集的周期。 采集周期如果是 ``0``
则代表这个插件是持续运行的，agent 不会设置超时时间将插件杀死。
插件输出的格式跟跟带参数插件格式一致（上面的 json）。

之后可以在规则中配置使用这个插件（假设你把这个插件放到了 ``plugin/nginx`` 目录）

.. code-block:: clojure

    (def some-nginx-related-rules
      (where (host "nginx-machine")
        (plugin-dir "nginx")
        (your-other-rules ...)))

持续输出的插件
--------------

插件可以输出多行 json，输出的 json 会马上被 agent 收集起来并上报。

此类插件可以将采集周期 ``step`` 设置成 ``0`` ，agent
会认为插件没有超时时间，并且只会在插件失败后才会重新调度。
