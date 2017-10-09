.. _filtering:

过滤事件
========

where
-----
``(where expr & children)``

按照 ``expr`` 指定的条件过滤事件，将符合条件事件向下传递。

请参考 `where 的官方文档 <http://riemann.io/api/riemann.streams.html#var-where>`_

by
--

``(by fields & children)``

将事件流按照 ``fields`` 指定的字段分离成多个子流，指定字段具有相同值的事件进入相同的子流，
每个子流会独立地进行剩下的处理过程。

请参考 `by 的官方文档 <http://riemann.io/api/riemann.streams.html#var-by>`_

runs
----

``(runs len-run field & children)``

当 ``field`` 指定的字段值在过去的 ``len-run`` 次没有变化时，将事件向下传递。

请参考 `runs 的官方文档 <http://riemann.io/api/riemann.streams.html#var-runs>`_

changed
-------

``(changed pred & children)``

当 ``(pred event)`` 的值与上次比较变化了，那么将这一次的事件向下传递。

请参考 `changed 的官方文档 <http://riemann.io/api/riemann.streams.html#var-changed>`_
