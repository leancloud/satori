.. _watchdog-streams:

看门狗
======

看门狗可以实现在指定的事件一定时间之后没有上报之后主动报警。


.. _feed-dog:

feed-dog
--------

``(feed-dog ttl)``

``(feed-dog ttl outstanding-tags)``


喂狗。当有事件流向这个流，看门狗就会静默。如果 ``ttl`` 秒之后没有再次喂狗，就会触发 watchdog 报警，之前缓存住的事件会在 :ref:`watchdog` 流后面出现，这时就可以报警了。

事件的区分与 alarm 的区分是一致的， ``outstanding-tags`` 的说明可以参照 :ref:`exclaimation-mark` 。


.. _watchdog:

watchdog
--------

``(watchdog & children)``

过滤出看门狗事件，用于之后的判定和报警。

在事件超时之后，超时的事件会用特殊的方式包装起来，需要用这个流来提取出来。

这个流需要直接接在最外层，即需要看到集群内所有的事件。会将解包的过期事件向下传递。

.. code-block:: clojure

  (sdo
    (where (service "agent.alive")
      (feed-dog 90))

    (watchdog
      (where (service "agent.alive")
        (! {:note "Agent.Alive 不上报了！"
            :level 2
            :expected 1
            :outstanding-tags [:region]
            :groups [:operation]})))

.. warning::
    ``watchdog`` 流下出现的事件都是已经过期或者恢复的报警，
    ``state`` 已经帮你设置成 ``:problem`` 或者 ``:ok`` 了，并且一个周期内只会出现一次，
    所以请不要接 ``(judge ...)`` 流或者 ``(runs 2 ...)`` 之类过滤的流，直接用 ``where``
    过滤出想要的结果喂给 :ref:`exclaimation-mark` 就可以了。
