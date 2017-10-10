.. _builtin-ss:

TCP 统计
========

.. note::
   这里的指标都是 ``ss -s`` 的输出

ss.estab
   :意义: 已经建立的 TCP Socket 数量
   :取值: 0 - 无上限，整数
   :Tags: 无

ss.closed
   :意义: 已经建立的 TCP Socket 数量
   :取值: 0 - 无上限，整数
   :Tags: 无

   .. note::
      或者是『处于 ``CLOSE_WAIT`` 、 ``LAST_ACK``、 ``TIME_WAIT`` 状态的 TCP Socket 数量』

ss.orphaned
   :意义: 没有用户态 fd 的 TCP Socket 数量
   :取值: 0 - 无上限，整数
   :Tags: 无

   .. note::
      或者是『处于 ``FIN_WAIT_1`` 、 ``FIN_WAIT_2``、 ``CLOSING`` 状态的 TCP Socket 数量』

      在用户态程序发送数据并关闭 fd，若此时这个 fd 关联的 socket
      还没有将缓冲区内的数据全部发出去，则在内核中会存在一个 orphaned socket。

ss.synrecv
   :意义: 处于 ``SYN_RECV`` 状态的 TCP Socket 数量
   :取值: 0 - 无上限，整数
   :Tags: 无

ss.timewait
   :意义: 处于 ``TIMEWAIT`` 状态的 TCP Socket 数量
   :取值: 0 - 无上限，整数
   :Tags: 无
