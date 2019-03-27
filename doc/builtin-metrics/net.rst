.. _builtin-net:

网络相关
========

.. warning::
   这里的表述模糊的指标具体什么情况会 +1 是网卡驱动程序决定的，需要看源码确定

net.if.in.bytes
    :意义: 网卡成功收到的字节数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.in.packets
    :意义: 网卡成功收到的包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.in.errors
    :意义: 网卡收到的错误包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.in.dropped
    :意义: 网络栈丢掉的包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

    .. note::
        这里统计的包已经被网卡接收到了，但是在处理过程中被丢掉了。

        比如内存不足无法分配 skb。

net.if.in.fifo.errs
    :意义: 因为接收缓冲区满丢掉的包
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.in.frame.errs
    :意义: 网卡接收到的错误帧数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.in.compressed
    :意义: 网卡接收到的压缩包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

    .. note::
        只对支持压缩的网卡有意义（比如 PPP 的网卡）

net.if.in.multicast
    :意义: 网卡收到的多播包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.out.bytes
    :意义: 网卡成功发送的字节数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.out.packets
    :意义: 网卡成功发送的包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.out.errors
    :意义: 网卡发送错误的包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.out.dropped
    :意义: 网卡丢弃的发送包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.out.fifo.errs
    :意义: 网卡发送错误的包数量(FIFO相关)
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.out.collisions
    :意义: 网卡发送包在链路上冲突的数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.out.carrier.errs
    :意义: 网卡发送包因为信号问题失败的数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.out.compressed
    :意义: 网卡发送的压缩过的包的数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

    .. note::
       只对支持压缩的网卡有意义（比如 PPP 的网卡）

net.if.total.bytes
    :意义: 网卡发送/接收的字节数总和
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.total.packets
    :意义: 网卡发送/接收的数据包总和
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.total.errors
    :意义: 网卡发送/接收的错误包总和
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}

net.if.total.dropped
    :意义: 网卡发送/接收时被丢弃的包总和
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"iface": "``网卡名``"}
