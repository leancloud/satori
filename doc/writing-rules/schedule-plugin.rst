.. _schedule-plugin:

调度插件
========

.. _plugin-dir:

plugin-dir
----------

``(plugin-dir & dirs)``

所有流过这个流的事件中的主机都会执行 ``dirs`` 指定的目录中的插件。

.. code-block:: clojure

    (where (host #"^redis-.*$")
        (plugin-dir "redis"))

这个例子中，当 ``redis-foo`` 主机的第一个事件到达这个流的时候，
就会被调度规则仓库中 ``plugin/redis`` 目录中的的插件。

.. note::
    因为 agent 总会发送 ``agent.alive`` 事件，所以不用担心插件无法调度的情况。


.. _plugin:

plugin
------

``(plugin metric step args)``

为机器指定一个接受参数的插件。
其中 ``metric`` 是插件名称，在规则仓库中的 :file:`plugin/<metric>` 中。
插件会每个 ``step`` 秒调度一次，并且将 ``args`` 的参数通过标准输入传递给插件。

.. code-block:: clojure

    (where (host "url-checker")
      (plugin "url.check" 30 {:name "example",
                              :url "http://example.com",
                              :timeout 15}))

这个例子中会在 ``url-checker`` 这个主机上每隔30秒执行一下 ``url.check`` 插件，指定的参数也会传递给插件。

具体参数是怎么传递的请参考 :ref:`writing-plugin` 。

.. note::
   在调度持续执行的插件时，这里的 ``step`` 须填 ``0``
