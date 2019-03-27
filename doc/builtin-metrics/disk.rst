.. _builtin-disk:

磁盘利用率
==========

计算值（对标 iostat 工具，比较容易使用）
----------------------------------------

disk.io.read\_bytes
    :意义: 磁盘读取的字节数（较上一次取样）
    :取值: 0 - 无上限，整数，单位：Bytes
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

disk.io.write\_bytes
    :意义: 磁盘写入的字节数（较上一次取样）
    :取值: 0 - 无上限，整数，单位：Bytes
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

disk.io.avgrq\_sz
    :意义: 平均 IO 大小
    :取值: 0 - 无上限，单位：Sectors/IORequest
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

disk.io.avgqu\_sz
    :意义: 平均 IO 队列长度
    :取值: 0 - 无上限，单位：Sectors
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

disk.io.await
    :意义: 平均 IO 耗时
    :取值: 0 - 无上限，单位：ms
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

disk.io.svctm
    `已废弃的指标，不要再使用了`


原始值（/proc/diskstat 的原始数据）
-----------------------------------

disk.io.read\_requests
    :意义: 磁盘完成的读请求数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

    .. note::

       合并的多个请求在这里算1个

disk.io.read\_merged
    :意义: 被合并的读请求数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

    .. note::

        在队列中的请求如果相邻会被合并成一个大的请求

disk.io.read\_sectors
    :意义: 磁盘完成读取的扇区数
    :取值: 0 - 无上限，整数，单调递增(COUNTER)
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

disk.io.msec\_read
    :意义: 磁盘花在读取上的总时间
    :取值: 0 - 无上限，整数，单位：毫秒(ms)，单调递增(COUNTER类型)
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

disk.io.write\_requests
    :意义: 磁盘完成的写请求数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

    .. note::

       合并的多个请求在这里算1个

disk.io.write\_merged
    :意义: 被合并的写请求数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

    .. note::

       在队列中的请求如果相邻会被合并成一个大的请求

disk.io.write\_sectors
    :意义: 磁盘完成写入的扇区数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

disk.io.msec\_write
    :意义: 磁盘花在写入上的总时间
    :取值: 0 - 无上限，整数，单位：毫秒(ms)，单调递增(COUNTER类型)
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

disk.io.ios\_in\_progress
    :意义: 当前正在处理的请求数
    :取值: 0 - 无上限，整数，单位：毫秒(ms)
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

    .. note::

       或者叫『排队中的请求数』

disk.io.msec\_total
    :意义: 磁盘花在处理请求上的总时间
    :取值: 0 - 无上限，整数，单位：毫秒(ms)，单调递增(COUNTER类型)
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

    .. note::

       或者叫『请求队列不为空的总时间』

disk.io.msec\_weighted\_total
    :意义: 磁盘花在处理请求上的总加权时间
    :取值: 0 - 无上限，整数，单位：毫秒(ms)，单调递增(COUNTER类型)
    :Tags: {"device": "``设备路径, 比如 /dev/sda``"}

    .. note::

       这里的值的意义是『所有请求的总等待时间』

       每一次请求结束后，这个值会增加这个请求的处理时间乘以当前的队列长度
