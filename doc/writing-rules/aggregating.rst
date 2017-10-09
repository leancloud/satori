.. _aggregating:

数据聚集
========

.. _aggregate-star:

aggregate*
----------

``(aggregate* f & children)``

接受事件的数组，变换成一个事件。 ``f`` 函数接受事件的值的数组作为参数，返回一个计算过后值，
然后将 ``aggregate`` 会将这个值替换掉最后一个事件中的值，并向下传递这个事件。

希望计算一组数据的和、平均、方差…… 等等的话就需要这个流。

常接在各种窗口后面。

.. code-block:: clojure

  (fixed-time-window 60
    (aggregate* #(apply + %)
      (set-state (> 100)
        (! ...))))


.. _aggregate:

aggregate
---------

``(aggregate f & children)``

与 :ref:`aggregate-star` 一样， 区别是 ``f`` 接受可变长参数，
而不是一个 vector（ ``(f & vals)`` 而不是 ``(f vals)`` ）

例子参见 :ref:`complex-rule-demo` 。


.. _to-difference:

->difference
------------

``(->difference & children)``

接受事件的数组，变换成另外一个事件数组。
新的事件数组中，每一个事件的监控值是之前相邻两个事件监控值的差。

如果你的事件是一个一直增长的计数器，那么用这个流可以将它变成每次实际增长的值。

.. code-block:: clojure

    (moving-time-window 60
      (->difference
        (aggregate maxpdiff
          (set-state-gapped (|>| 0.5) (|<| 0.1)
            (! ...)))))


.. _avgpdiff:

avgpdiff
--------

``(avgpdiff & m)``

计算最后一个点相比于之前的点的平均值的变化率。
与 :ref:`aggregate` 搭配使用。

.. _maxpdiff:

maxpdiff
--------

``(maxpdiff & m)``

计算列表中最后一个点相对之前的点的最大变化率(MAX Percentage DIFFerence)，
与 :ref:`aggregate` 搭配使用。

计算变化率时总是使用两个点中的最小值做分母，
所以由1变到2的变化率是 1.0，
由2变到1的变化率是 -1.0 （而不是 -0.5)


.. _abs-compare:

\|>\| 和 \|<\|
--------------

``(|>| & args)``
``(|<| & args)``

将 ``args`` 取绝对值后进行比较。
通常你不会关心变化率的符号（变化方向），只对绝对值感兴趣。
在 :ref:`set-state` 和 :ref:`set-state-gapped` 上可以用这个函数作比较。

参见 :ref:`to-difference` 上的例子。


.. _riemann-windows:

Riemann 自带的窗口函数
----------------------

``(fixed-event-window n & children)``
``(fixed-time-window n & children)``
``(moving-event-window n & children)``
``(moving-time-window n & children)``

窗口函数会收集传下来的事件并缓存住，然后按照指定的规则将缓存住的事件以数组（vector）向下传递。

+--------------+------------------------------------------------------------------------------------+
| 类型         | 解释                                                                               |
+==============+====================================================================================+
| fixed event  | 收集过去 ``n`` 个事件后向下传递，然后再收集下一组 ``n`` 个事件                     |
+--------------+------------------------------------------------------------------------------------+
| moving event | 维护一个 ``n`` 个事件的滑动窗口，每到达一个新事件后将过去 ``n`` 个事件向下传递。   |
+--------------+------------------------------------------------------------------------------------+
| fixed time   | 收集过去 ``n`` 秒内的事件，超时后向下传递，然后再收集下一组 ``n`` 秒内收集的事件。 |
+--------------+------------------------------------------------------------------------------------+
| moving time  | 维护一个 ``n`` 秒的滑动窗口，每到达一个新事件将过去 ``n`` 秒内的事件向下传递       |
+--------------+------------------------------------------------------------------------------------+


参考 :ref:`to-difference` 中的代码样例

另外可参考 `Riemann 官方文档 <http://riemann.io/api/riemann.streams.html#var-moving-event-window>`_ 。

.. _group-window:

group-window
------------

``(group-window group-fn & children)``

将事件分组后向下传递，类似 :ref:`riemann-windows` ，但不使用时间或者事件个数进行切割，
而是通过 ``(group-fn event)`` 的值进行切割。 ``(group-fn event)`` 的值会被记录下来，
每一次出现重复值的时候，会将当前缓存住的事件数组向下传递。

比如你有一组同质的机器，跑了相同的服务，但是机器名不一样，可以通过

.. code-block:: clojure

  (group-window :host
    ...)

将事件分组后处理（e.g. 对单台的容量求和获得总体容量）

举例：一个事件流中的事件先后到达，其中 ``:host`` 的值如下

.. code-block:: text

    a b c d b a c a b

那么会被这个流分成

.. code-block:: text

    [a b c d] [b a c]

分成 2 次向下游传递，最后 的 ``[a b]`` 因为还没有重复事件到达所以还会在缓冲区内等待。


.. _slot-window:

slot-window
-----------

``(slot-window slot-fn fields & children)``

收集指定的几个事件并打包向下传递。事件的特征由 ``(slot-fn event)`` 提取，
并与 ``fields`` 中的的定义匹配，如果 ``fields`` 中的所有条件匹配的事件都收集到了，
则打包向下传递并开始下一轮收集。

``fields`` 是形如

.. code-block:: clojure

  {:key1 "value1", :key2 "value2"}

的 map， ``:key1`` 和 ``:key2`` 是自己定义的，用于引用匹配后的事件，
``"value1"`` 和 ``"value2"`` 是希望匹配的 ``(slot-fn event)`` 的值，与之相等的事件会被放到相应的槽中。

``slot-window`` 会向下传递形如

.. code-block:: clojure

   {:key1 event1, :key2 event2}

的 map，其中 ``event1`` 和 ``event2`` 是匹配到的事件。

与 :ref:`group-window` 相反，:ref:`group-window` 收集一组同质的事件，
``slot-window`` 用于收集一组异质的事件。

当 ``slot-window`` 遇到重复的事件但是还没有满足向下传递的条件时，新的事件会替换掉缓存住的已有事件。

常用于收集同一个资源不同的指标用于复杂的判定。

比如有一个服务，同时收集了错误请求量和总量，希望按照错误数量在一定之上后按照错误率报警

.. code-block:: clojure

  (slot-window :service {:error "app.req.error"
                         :count "app.req.count"}

    ; 此时会有形如 {:error {:service "app.req.error", ...},
    ;               :count {:service "app.req.count", ...}} 的事件传递下来

    ; 构造出想要的 event
    (slot-coalesce {:service "app.req.error_rate"
                    :metric (if (> error 100) (/ error count) -1)}
      (set-state (> 0.5)
        (runs :state 5
          (should-alarm-every 120
            (! ...))))))

.. _slot-coalesce:

slot-coalesce
-------------

``(slot-coalesce ev' & children)``

对 :ref:`slot-window` 的结果进行计算，并构造出单一的事件。

``ev'`` 是构造出的新事件的模板，这里的值会被复制到新事件中，
并且在模板中可以直接引用 :ref:`slot-window` 中定义的事件的值。

以 :ref:`slot-window` 中的代码为例，表达式中可以直接用如下的约定引用槽中的值：

+------------------------------+------------------------------------------------+
| 变量名                       | 意义                                           |
+==============================+================================================+
| ``error`` ， ``count``       | ``slot-window`` 中定义的槽捕捉到的事件的监控值 |
+------------------------------+------------------------------------------------+
| ``ev:error`` ， ``ev:count`` | ``slot-window`` 中定义的槽捕捉到的事件本身     |
+------------------------------+------------------------------------------------+
| ``event``                    | ``slot-window`` 传递下来的 map 本身            |
+------------------------------+------------------------------------------------+
