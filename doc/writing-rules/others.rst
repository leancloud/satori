.. _other-streams:

其他需要用到的流
================

.. _smap:

smap
----
``(smap f & children)``

将事件 ``event`` 变换成 ``(f event)`` 后，若变换后的事件不是 ``nil`` ，则向下传递。

请参考 `smap 官方文档 <http://riemann.io/api/riemann.streams.html#var-smap>`_


.. _copy:

copy
----

``(copy from to & children)``

将事件的 ``from`` 字段复制到 ``to`` 字段。
通常接在 :ref:`aggregate` 后面，用于修正 ``:host`` 。

.. code-block:: clojure

    (aggregate +
      (copy :region :host
        (...))

.. _sdo:

sdo
---

``(sdo & children)``

将多个子流合并成一个流，传下来的事件会给每一个子流派发。

因为 ``def`` 后面只能接一个 form， 如果需要将多个流接在同一个 ``def`` 后面，就需要 ``sdo`` 来包裹一下。

.. code-block:: clojure

    (sdo
        (rule1 ...)
        (rule2 ...))

.. _waterfall:

->waterfall
-----------

``(->waterfall & children)``

用这个将规则改写成瀑布流的方式，会比较好看。
仅适用于每个规则下仅有一个子流的情况。

比如：

.. code-block:: clojure

    (->waterfall
      (where (service "nvgpu.gtemp"))
      (by :host)
      (copy :id :aggregate-desc-key)
      (group-window :id)
      (aggregate max)
      (judge-gapped (> 90) (< 86))
      (alarm-every 2 :min)
      (! {:note "GPU 过热"
          :level 1
          :expected 85
          :outstanding-tags [:host]
          :groups [:operation :devs]}))

与如下代码等价：

.. code-block:: clojure

    (where (service "nvgpu.gtemp")
      (by :host
        (copy :id :aggregate-desc-key
          (group-window :id
            (aggregate max
              (judge-gapped (> 90) (< 86)
                (alarm-every 2 :min
                  (! {:note "GPU 过热"
                      :level 1
                      :expected 85
                      :outstanding-tags [:host]
                      :groups [:operation :devs]}))))))))

.. note::

   并没有推荐这样做，要不要看个人喜好就好。
