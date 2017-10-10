.. _builtin-agent:

Agent 存活
==========

agent
   :意义: Agent 存活
   :取值: 1
   :Tags: 无

   .. warning::
      agent 在不存活的时候并不会上报 0，而是不上报。

      检测 Agent 存活推荐用 :ref:`url-check` 插件主动探测 agent 的
      ``/v1/ping`` 接口。

      一定要用这个的话，可以看一下 :ref:`watchdog` 流
