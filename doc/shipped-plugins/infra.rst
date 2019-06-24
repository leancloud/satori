.. _infra:

基础监控插件
============

satori-agent 内置指标的扩展，建议每个机器上都收集

插件文件地址
    infra/*

插件类型
    不接受参数（直接执行），重复执行


.. _infra-conntrack:

Conntrack
---------

Conntrack 是 Netfilter 用来跟踪连接的机制，如果使用了 NAT 或者有状态防火墙功能就会启用 conntrack 跟踪连接。

net.netfilter.conntrack.used
    :意义: Conntrack 桶用量
    :提供: :file:`infra/30_conntrack.py`
    :取值: 0 - 无上限，整数
    :Tags: 无

net.netfilter.conntrack.max
    :意义: Conntrack 桶最大值
    :提供: :file:`infra/30_conntrack.py`
    :取值: 0 - 无上限，整数
    :Tags: 无

net.netfilter.conntrack.used_ratio
    :意义: Conntrack 桶用量百分比
    :提供: :file:`infra/30_conntrack.py`
    :取值: 0 - 1.0，浮点数
    :Tags: 无


.. _infra-normalized-load:

重整化的机器负载（Load）
------------------------

load.1min.normalized
   :意义: 1分钟平均负载
   :提供: :file:`infra/30_normalized_load.py`
   :取值: 0 - 无上限，浮点数
   :Tags: 无

load.5min.normalized
   :意义: 5分钟平均负载
   :提供: :file:`infra/30_normalized_load.py`
   :取值: 0 - 无上限，浮点数
   :Tags: 无

load.15min.normalized
   :意义: 15分钟平均负载
   :提供: :file:`infra/30_normalized_load.py`
   :取值: 0 - 无上限，浮点数
   :Tags: 无

.. note::
    这个值与 :ref:`builtin-load` 监控的是相同的值，
    但是这里的 Load 会除以 CPU 核数，在编写监控规则的时候就不用考虑核数了。


.. _infra-megaraid:

MegaRAID 磁盘损坏监控
---------------------

megaraid.offline
    :意义: 第一个 MegaRAID 控制器中 Failed/Offline 状态的磁盘个数
    :提供: :file:`infra/600_megaraid.py`
    :取值: 0 - 无上限，整数
    :Tags: 无

    .. note::
       这个插件需要安装 ``megacli`` 工具后才能使用，
       请确认 :file:`/opt/MegaRAID/MegaCli/MegaCli64` 文件是否存在。


.. _infra-kernel:

内核 dmesg 监控
---------------

kernel.dmesg.bug
    :意义: dmesg 中出现 ``BUG:`` 字样的次数
    :提供: :file:`infra/60_kernel.py`
    :取值: 0 - 无上限，整数
    :Tags: 无

kernel.dmesg.io_error
    :意义: dmesg 中出现 ``I/O error`` 字样的次数
    :提供: :file:`infra/60_kernel.py`
    :取值: 0 - 无上限，整数
    :Tags: 无


.. note::
    处理完故障之后可以使用 ``dmesg --clear`` 来清除内核的 dmesg buffer，否则会一直报告 0 以上的值。


.. _infra-zombies:

异常状态进程数量
----------------
proc.zombies
    :意义: 僵尸进程数量
    :提供: :file:`infra/60_zombies.py`
    :取值: 0 - 无上限，整数
    :Tags: 无

    .. note::
       僵尸进程是已经结束的进程，不占用内存/CPU，只占用内核进程表的一个条目。
       大量出现僵尸进程通常是程序编写的问题。

proc.uninterruptables
    :意义: 处于不可中断睡眠状态的进程数量
    :提供: :file:`infra/60_zombies.py`
    :取值: 0 - 无上限，整数
    :Tags: 无

    .. note::
       最常见的不可中断睡眠状态的进程状态是由磁盘 IO 导致的，是正常状态。
       如果持续性的数量过多就需要调查了。

.. _infra-softirq:

软中断统计
----------
softirq.timer
    :意义: 时钟中断
    :提供: :file:`infra/30_softirq.py`
    :取值: 0 - 无上限，累积，整数
    :Tags: 无

softirq.net_tx: 89 84
    :意义: 发送网络数据
    :提供: :file:`infra/30_softirq.py`
    :取值: 0 - 无上限，累积，整数
    :Tags: 无

softirq.net_rx: 776 1331
    :意义: 接收网络数据
    :提供: :file:`infra/30_softirq.py`
    :取值: 0 - 无上限，累积，整数
    :Tags: 无

softirq.block: 60 89
    :意义: 块设备请求
    :提供: :file:`infra/30_softirq.py`
    :取值: 0 - 无上限，累积，整数
    :Tags: 无

softirq.tasklet: 141 29
    :意义: Tasklet
    :提供: :file:`infra/30_softirq.py`
    :取值: 0 - 无上限，累积，整数
    :Tags: 无

    .. note::
       一些内核中高优先级但是不合适做成软中断的任务

softirq.sched:
    :意义: 进程调度
    :提供: :file:`infra/30_softirq.py`
    :取值: 0 - 无上限，累积，整数
    :Tags: 无

softirq.rcu:
    :意义: RCU 回收
    :提供: :file:`infra/30_softirq.py`
    :取值: 0 - 无上限，累积，整数
    :Tags: 无

   .. note::
      RCU 是内核中用的一种并发数据结构，需要定期清理


监控规则样例
------------

.. note::
   请直接参考规则仓库中附带的 :file:`infra/common.clj` 文件，篇幅过长不再贴了。
