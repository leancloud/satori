.. _judging-and-alarming:

判定和发送报警
==============

.. _exclaimation-mark:

发送报警（!）
-------------

``(! m)``

创建一个发送报警的流，流经这个流的事件都会被发到 alarm 产生报警。
接受一个 map 做参数，map 中需要可以指定如下的参数：

+-------------------+--------+-------------------------+
| Key               | 类型   | 意义                    |
+===================+========+=========================+
| :note             | string | 报警标题 [#]_           |
+-------------------+--------+-------------------------+
| :level            | int    | 报警级别 [#]_           |
+-------------------+--------+-------------------------+
| :event?           | bool   | 事件类型？[#]_          |
+-------------------+--------+-------------------------+
| :expected         | float  | 期望的正常值 [#]_       |
+-------------------+--------+-------------------------+
| :outstanding-tags | vector | 区分报警的 tags [#]_    |
+-------------------+--------+-------------------------+
| :groups           | vector | 将报警发送到报警组 [#]_ |
+-------------------+--------+-------------------------+

.. [#] 标题对于一个特定的报警是不能变的，不要把报警的数据编码在这里面！
.. [#] 约定 0 级别最高，最小是15，不过一般来说用不到那么多级别。
       报警级别影响报警方式。
.. [#] 期望的正常值。 这个值暂时没有用到，但是也最能填上。
.. [#] **可选** ，默认是 ``false`` 。事件类型只会发送报警，不会记录和维护状态，无法在 alarm 中看到。
.. [#] **可选** ，默认为事件中所有的 tag。在这里指定的 tag 值的组合如果不一样，
       就会被 alarm 当做不同的报警分别追踪
.. [#] 报警组的配置请参考 :ref:`alarm-config-users-and-groups`


样例：

.. code-block:: clojure

    (set-state (> 1) ; 假设
      (! {:note "服务器炸了！"
          :level 1
          :event? false
          :expected 0
          :outstanding-tags [:region]
          :groups [:operation :boss]}))

.. note::
    ``!`` 接受的事件需要将 ``:state`` 设置成 ``:problem`` 或者 ``:ok`` 来表示是有问题还是恢复。

    参见下文的 :ref:`set-state` 和 :ref:`set-state-gapped`


.. _set-state-star:

set-state*
----------

``(set-state* c & children)``

设置事件的状态。 ``c`` 是接受事件作为参数的函数。
``c`` 返回值为 ``true`` 则会将事件的 ``:state`` 设置成 ``:problem`` ，否则会设置成 ``:ok``

.. code-block:: clojure

    (set-state* #(> (:metric %) 1)
      (! ...))


.. _set-state:

set-state
---------

``(set-state c & children)``

参见 :ref:`set-state-star` ，这里 c 是形如 ``(> 1.0)`` 的 form。


.. code-block:: clojure

    (set-state (> 1)
      (! ...))

.. note::
    ``set-state (> 1.0) ...)`` 会被重写成 ``(set-state* #(> (:metric %) 1.0) ...)``


.. _set-state-gapped-star:

set-state-gapped*
-----------------

``(set-state-gapped* rising falling & children)``

设置事件的状态。 ``rising`` 是 OK -> PROBLEM 的条件， ``falling`` 是 PROBLEM -> OK 的条件

参见 :ref:`set-state-star`

.. code-block:: clojure

    (set-state-gapped* #(> (:metric %) 10) #(< (:metric %) 1)
      (! ...))


.. _set-state-gapped:

set-state-gapped
----------------

``(set-state-gapped rising falling & children)``

参见 :ref:`set-state-gapped-star` ，这里 ``rising`` 和 ``falling`` 是形如 ``(> 1.0)`` 的 form。


.. code-block:: clojure

    (set-state-gapped (> 10) (< 1)
      (! ...))


.. _should-alarm-every:

should-alarm-every
------------------

``(should-alarm-every dt & children)``

用于对报警事件限流，通常接在 :ref:`exclaimation-mark` 流前面。
当事件的 ``:state`` 是 ``:problem`` 时，每 ``dt`` 秒向下传递一次。
当事件的 ``:state`` 由 ``:problem`` 变成 ``:ok`` 时，向下传递一次。
其他时间不放行。

.. code-block:: clojure

    (set-state (> 1)
      (should-alarm-every 60
        (! ...)))
