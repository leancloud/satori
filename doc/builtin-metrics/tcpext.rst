.. _builtin-tcpext:

高级 TCP 指标
=============

.. note::
   这里的文档参(chao)考(xi)了很多 http://perthcharles.github.io/2015/11/10/wiki-netstat-proc
   和 http://www.cnblogs.com/lovemyspring/articles/5087895.html 的内容

   这里的大部分指标可以通过 ``netstat -st`` 命令获得

数据包统计
----------
TcpExt.EmbryonicRsts
    :意义: 在 ``SYN_RECV`` 状态收到带 RST/SYN 标记的包个数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

SYN Cookies 功能
------------------
From Wikipedia:

    SYN cookie 是一种用于阻止 SYN flood 攻击的技术。
    这项技术的主要发明人 Daniel J. Bernstein 将 SYN cookies 定义为“TCP 服务器进行的对开始 TCP 数据包序列数字的特定选择”。
    举例来说，SYN Cookies 的应用允许服务器当 SYN 队列被填满时避免丢弃连接。
    相反，服务器会表现得像 SYN 队列扩大了一样。服务器会返回适当的 SYN+ACK 响应，但会丢弃 SYN 队列条目。
    如果服务器接收到客户端随后的 ACK 响应，服务器能够使用编码在 TCP 序号内的信息重构 SYN 队列条目。

SYN Cookies 一般不会被触发，只有在 ``tcp_max_syn_backlog`` 队列被占满时才会被触发，
因此 ``SyncookiesSent`` 和 ``SyncookiesRecv`` 一般应该是0。
但是 ``SyncookiesFailed`` 值即使SYN Cookies 机制没有被触发，也很可能不为0。
这是因为一个处于 ``LISTEN`` 状态的 socket 收到一个不带 SYN 标记的数据包时，就会调
用 ``cookie_v4_check()`` 尝试验证 cookie 信息。
而如果验证失败，``SyncookiesFailed`` 次数就加1。

TcpExt.SyncookiesSent
    :意义: 使用 SYN Cookie 发送的 SYN/ACK 包个数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.SyncookiesRecv
    :意义: 收到携带有效 SYN Cookie 信息的包的个数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.SyncookiesFailed
    :意义: 收到携带无效 SYN Cookie 信息的包的个数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TIME_WAIT 回收
------------------
``TIME_WAIT`` 状态是 TCP 协议状态机中的重要一环，服务器设备一般都有非常多处于 ``TIME_WAIT`` 状态的 socket。
如果是在主要提供 HTTP 服务的设备上，``TW`` 值应该接近 ``TcpPassiveOpens`` 值。
一般情况下，``sysctl_tcp_tw_reuse`` 和 ``sysctl_tcp_tw_recycle`` 都是不推荐开启的， `这里解释了为什么`_ 。
所以 ``TWKilled`` 和 ``TWRecycled`` 都应该是0。
同时 ``TCPTimeWaitOverflow`` 也应该是0，否则就意味着内存使用方面出了大问题。

TcpExt.TW
    :意义: 经过正常的的超时结束 ``TIME_WAIT`` 状态的 socket 数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TWRecycled
    :意义: 通过 ``tcp_tw_reuse`` 机制结束 ``TIME_WAIT`` 状态的 socket 数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
       注意这里命名的不一致……

       这是 Linux 内核的坑，所以保留下来了

TcpExt.TWKilled
    :意义: 通过 ``tcp_tw_recycle`` 机制结束 ``TIME_WAIT`` 状态的 socket 数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. warning::
       你不应该关心这个值， ``tcp_tw_recycle`` 机制即将被取消。

    .. note::
       注意这里命名的不一致……

       这是 Linux 内核的坑，所以保留下来了

       只有在 ``net.ipv4.tcp_tw_recycle`` 开启时，这个值才会增加。

TcpExt.TCPTimeWaitOverflow
    :意义: 因为超过限制而无法分配的 ``TIME_WAIT`` socket 数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        这个数量可以在 ``net.ipv4.tcp_max_tw_buckets`` 调整

.. _`这里解释了为什么`: http://perthcharles.github.io/2015/08/27/timestamp-NAT/

超时重传相关
------------------
From Wikipedia:

    Whenever a packet is sent, the sender sets a timer that is a conservative estimate of when that packet will be acked. If the sender does not receive an ack by then, it transmits that packet again. The timer is reset every time the sender receives an acknowledgement. This means that the retransmit timer fires only when the sender has received no acknowledgement for a long time. Typically the timer value is set to :math:`{\displaystyle {\text{smoothed RTT}}+\max(G,4\times {\text{RTT variation}})}` where :math:`{\displaystyle G}` is the clock granularity. Further, in case a retransmit timer has fired and still no acknowledgement is received, the next timer is set to twice the previous value (up to a certain threshold). Among other things, this helps defend against a man-in-the-middle denial of service attack that tries to fool the sender into making so many retransmissions that the receiver is overwhelmed.

RTO 超时对 TCP 性能的影响是巨大的，因此关心 RTO 超时的次数也非常必要。

TcpExt.TCPTimeouts
    :意义: RTO timer第一次超时的次数，仅包含直接超时的情况
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::

        3.10 中的 TLP 机制能够减少一定量的 ``TCPTimeouts`` 数，将其转换为快速重传。
        关于TLP的原理部分，可参考 `这篇wiki`_ 。

TcpExt.TCPSpuriousRTOs
    :意义: 通过F-RTO机制发现的虚假超时个数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPLossProbes
    :意义: Probe Timeout(PTO) 导致发送 Tail Loss Probe (TLP) 包的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPLossProbeRecovery
    :意义: 丢失包刚好被TLP探测包修复的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPRenoRecovery:
    :意义: 进入 Recovery 阶段的次数，对端不支持 SACK 选项
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPSackRecovery:
    :意义: 进入 Recovery 阶段的次数，对端支持 SACK 选项
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPRenoRecoveryFail:
    :意义: 先进入 Recovery 阶段，然后又 RTO 的次数，对端不支持 SACK 选项
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPSackRecoveryFail:
    :意义: 先进入 Recovery 阶段，然后又 RTO 的次数，对端支持 SACK 选项
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPRenoFailures
    :意义: 先进入 TCP_CA_Disorder 阶段，然后又 RTO 超时的次数，对端不支持 SACK 选项
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPSackFailures
    :意义: 先进入 TCP_CA_Disorder 阶段，然后又 RTO 超时的次数，对端支持 SACK 选项
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPLossFailures
    :意义: 先进入 TCP_CA_Loss 阶段，然后又 RTO 超时的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPSACKReneging
    :意义: 收到的不正常的 SACK 包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无


.. _`这篇wiki`: http://perthcharles.github.io/2015/10/31/wiki-network-tcp-tlp/


重传数量
------------------
TcpExt.TCPFastRetrans
    :意义: 成功快速重传的 skb 数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPForwardRetrans
    :意义: 成功 ForwardRetrans 的 skb 数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
       Forward Retrans 重传的序号高于 retransmit_high 的数据
       retransmit_high 目前的理解是被标记为 lost 的 skb 中，最大的 end_seq 值

TcpExt.TCPSlowStartRetrans
    :意义: 成功在Loss状态发送的重传 skb 数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        这里仅记录非 RTO 超时进入 Loss 状态下的重传数量。

        目前找到的一种非 RTO 进入 Loss 状态的情况是：``tcp_check_sack_reneging()`` 函数发现
        接收端违反(renege)了之前的 SACK 信息时，会进入 Loss 状态。

TcpExt.TCPLostRetransmit
    :意义: 丢失的重传 skb 数量，没有 TSO(TCP Segment Offloading) 时，等于丢失的重传包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPRetransFail
    :意义: 尝试 FastRetrans、ForwardRetrans、SlowStartRetrans 重传失败的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

.. note::
    这些计数器统计的重传包，都 **不是** 由于 RTO 超时导致进行的重传。

FastOpen
------------------
`TCP FastOpen(TFO)`_ 是 Google 提出来减少三次握手开销的技术，
核心原理就是在第一次建连时服务端计算一个 cookie 发给 client，之后客户端向
服务端再次发起建连请求时就可以携带 cookie ，如果 cookie 验证通过，
服务端可以不等三次握手的最后一个 ACK 包就将客户端放在 SYN 包里面的数据传递给应用层。

在 3.10 内核中，TFO 由 ``net.ipv4.tcp_fastopen`` 开关控制，默认值为0(关闭)。
而且 ``net.ipv4.tcp_fastopen`` 目前也是推荐关闭的，
因为网络中有些中间节点会丢弃那些带有不认识的 option 的 SYN 包。
所以正常情况下这些值也应该都是0，当然如果收到过某些不怀好意带 TFO cookie 信息的 SYN 包，
``TCPFastOpenPassive`` 计数器就可能不为0。

TcpExt.TCPFastOpenActive
    :意义: 发送的带 TFO cookie 的 SYN 包个数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPFastOpenPassive
    :意义: 收到的带 TFO cookie 的 SYN 包个数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPFastOpenPassiveFail
    :意义: 使用TFO技术建连失败的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPFastOpenListenOverflow
    :意义: TFO 请求数超过 listener queue 设置的上限则加1
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPFastOpenCookieReqd
    :意义: 收到一个请求 TFO cookie 的 SYN 包时加1
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

.. _`TCP FastOpen(TFO)`: https://tools.ietf.org/html/rfc7413

MD5
------------------
TCP MD5 Signature 选项是为提高 BGP Session 的安全性而提出的，详见 `RFC 2385`_ 。
因此内核中是以编译选项，而不是 sysctl 接口来配置是否使用该功能的。
如果内核编译是的 ``CONFIG_TCP_MD5SIG`` 选项未配置，则不会支持 TCPMD5Sig，下面两个计数器也就只能是0

TcpExt.TCPMD5NotFound
    :意义: 希望收到带MD5选项的包，但是包里面没有MD5选项
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPMD5Unexpected
    :意义: 不希望收到带MD5选项的包，但是包里面有MD5选项
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无


.. _`RFC 2385`: https://tools.ietf.org/html/rfc2385

DelayedACK
------------------
From Wikipedia:

    TCP delayed acknowledgment is a technique used by some implementations of the Transmission Control Protocol in an effort to improve network performance.
    In essence, several ACK responses may be combined together into a single response, reducing protocol overhead. However, in some circumstances, the technique can reduce application performance.

Delayed ACK 是内核中默认支持的，但即使使用 Delayed ACK，每收到两个数据包也
必须发送一个ACK。所以 ``DelayedACKs`` 可以估算为发送出去的 ACK 数量的一半。
同时 ``DelayedACKLocked`` 反应的是应用与内核争抢 socket 的次数，
如果占 ``DelayedACKs`` 比例过大可能就需要看看应用程序是否有问题了。

TcpExt.DelayedACKs
    :意义: 尝试发送 delayed ACK 的次数，包括未成功发送的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.DelayedACKLocked
    :意义: 由于用户态进程锁住了 socket，而无法发送 delayed ACK 的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.DelayedACKLost
    :意义: *TODO暂时不理解准确含义*
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPSchedulerFailed
    :意义: 如果在 delay ack 处理函数中发现 prequeue 还有数据，就加1。
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        `这里 <http://oss.sgi.com/archives/netdev/2002-10/msg01001.html>`_ 有一个看起来靠谱的解释:

            Its the number of packets that are prequeued (partially completed
            tcp processing, waiting for a recvmsg to complete & send ack etc)
            when the delayed ack timer goes off. i.e I think, the thinking is
            this shouldnt happen, since the delayed ack timer really shouldnt
            go off - the receiver should have picked up the skb's and sent
            the acks. I'm probably off by a mile here..Dont know.

        这个值应该非常接近于零才正常

DSACK
------------------
该类型计数器统计的是收/发 DSACK 信息次数。
``DSACKOldSent`` + ``DSACKOfoSent`` 可以当做是发送出的 DSACK 信息的次数，
而且概率上来讲 OldSent 应该占比更大。
同理 DSACKRecv 的数量也应该远多于 DSACKOfoRecv 的数量。
另外 DSACK 信息的发送是需要 ``net.ipv4.tcp_dsack`` 开启的，如果发现 sent 两个计数器为零，则要检查一下了。
一般还是建议开启 DSACK。

TcpExt.TCPDSACKOldSent
    :意义: 如果收到的重复数据包序号比 ``rcv_nxt`` 小，则+1
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        ``rcv_nxt`` 代表接收端想收到的下一个序号

TcpExt.TCPDSACKOfoSent
    :意义: 如果收到的重复数据包序号比 ``rcv_nxt`` 大，则是一个乱序的重复数据包，则+1
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPDSACKRecv
    :意义: 收到的 old DSACK 信息次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        判断 old 的方法：DSACK 序号小于 ACK 号

TcpExt.TCPDSACKOfoRecv
    :意义: 收到的 Ofo DSACK 信息次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        Ofo 是 Out-of-Order 的意思

TcpExt.TCPDSACKIgnoredOld
    :意义: 当一个 DSACK block 被判定为无效，且设置过 undo_marker，则+1
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TcpExt.TCPDSACKIgnoredNoUndo
    :意义: 当一个 DSACK block 被判定为无效，且未设置 undo_marker，则+1
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

.. note::
    关于 partial ACK 的完整内容可参考 `RFC6582`_ ，这里摘要定义部分

        In the case of multiple packets dropped from a single window of data,
        the first new information available to the sender comes when the
        sender receives an acknowledgment for the retransmitted packet (that
        is, the packet retransmitted when fast retransmit was first entered).
        If there is a single packet drop and no reordering, then the
        acknowledgment for this packet will acknowledge all of the packets
        transmitted before fast retransmit was entered.  However, if there
        are multiple packet drops, then the acknowledgment for the
        retransmitted packet will acknowledge some but not all of the packets
        transmitted before the fast retransmit.  We call this acknowledgment
        a partial acknowledgment.

.. _`RFC6582`: https://tools.ietf.org/html/rfc6582


Reorder
------------------
当发现了需要更新某条 TCP 流的 reordering 值(乱序值)时，以下计数器可能被使用到。
不过下面四个计数器为互斥关系，最少见的应该是 ``TCPRenoReorder`` ，毕竟 SACK 已经被
广泛部署使用了。

TcpExt.TCPFACKReorder
    :意义: 使用 FACK 机制检测到的乱序次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        与TCPSACKReorder类似，如果同时启用了 SACK 和 FACK，就增加本计数器。

TcpExt.TCPSACKReorder
    :意义: 使用 SACK 机制检测到的乱序次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        在 ``tcp_update_reordering()`` 中更新，当 ``metric > tp->reordering`` 并且启用 SACK 但关闭 FACK 时，本计数器加1

TcpExt.TCPRenoReorder
    :意义: 使用 Reno 快速重传机制检测到的乱序次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        在 ``tcp_update_reordering()`` 中更新，当 ``metric > tp->reordering`` 并且没有启用 SACK ，本计数器加1


TcpExt.TCPTSReorder
    :意义: 使用 TCP Timestamp 机制检测到的乱序次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_ack() -> tcp_fastretrans_alert() -> tcp_undo_partial() -> tcp_update_reordering()``

        Recovery 状态时，接收到到部分确认(snd_una < high_seq)时但已经 undo 完成(undo_retrans == 0)的次数。 数量上与 TCPPartialUndo 相等。


连接终止
--------

TCPAbortOnClose:
    :意义: 用户态程序在缓冲区内还有数据时关闭 socket 的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        此时会主动发送一个 RST 包给对端

TCPAbortOnData:
    :意义: socket 收到未知数据导致被关闭的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        如果在 ``FIN_WAIT_1`` 和 ``FIN_WAIT_2`` 状态下收到后续数据，或 ``TCP_LINGER2`` 设置小于0，则计数器加1

TCPAbortOnTimeout:
    :意义: 因各种计时器(RTO/PTO/keepalive)的重传次数超过上限而关闭连接的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

TCPAbortOnMemory:
    :意义: 因内存问题关闭连接的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        如果 orphaned socket 数量或者 tcp_memory_allocated 超过上限，则加1，
        一般值为0。

TCPAbortOnLinger:
    :意义: TCPAbortOnLinger
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        ``tcp_close()`` 中，因 ``tp->linger2`` 被设置小于 0，导致 ``FIN_WAIT_2`` 立即切换到 ``CLOSE`` 状态的次数，一般值为0。

TCPAbortFailed:
    :意义: 尝试结束连接失败的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        如果在准备发送 RST 时，分配 skb 或者发送 skb 失败，则加1，
        一般值为0

内存 Prune
----------
TcpExt.TCPMemoryPressures
    :意义: TCP 内存压力由“非压力状态”切换到“有压力状态”的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
       相关函数 ``tcp_enter_memory_pressure()``

       可能的触发点有:

       - tcp_sendmsg()
       - tcp_sendpage()
       - tcp_fragment()
       - tso_fragment()
       - tcp_mtu_probe()
       - tcp_data_queue()

TcpExt.PruneCalled
    :意义: 因为 socket 缓冲区满而被 prune 的包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_data_queue() -> tcp_try_rmem_schedule()``

        慢速路径中，如果不能将数据直接复制到用户态内存，需要加入到 sk_receive_queue 前，会检查 receiver side memory 是否允许，如果 rcv_buf 不足就可能 prune ofo queue。此时计数器加1。

TcpExt.RcvPruned
    :意义: RcvPruned
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_data_queue() -> tcp_try_rmem_schedule()``

        慢速路径中，如果不能将数据直接复制到用户态内存，需要加入到 sk_receive_queue 前，会检查 receiver side memory 是否允许，如果 rcv_buf 不足就可能 prune receive queue ，如果 prune 失败了，此计数器加1。

TcpExt.OfoPruned
    :意义:
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_data_queue() -> tcp_try_rmem_schedule()``

        慢速路径中，如果不能将数据直接复制到用户态内存，需要加入到 sk_receive_queue 前，会检查 receiver side memory 是否允许，如果 rcv_buf 不足就可能 prune ofo queue ，此计数器加1。


PAWS [#]_ 相关
--------------

.. [#] Protect Against Wrapping Sequence，TCP 序列号溢出保护

TcpExt.PAWSPassive
    :意义: 三路握手最后一个 ACK 的 PAWS 检查失败次数。
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_v4_conn_request()``

TcpExt.PAWSActive
    :意义: 在发送 SYN 后，接收到 ACK，但 PAWS 检查失败的次数。
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_rcv_synsent_state_process()``


TcpExt.PAWSEstab
    :意义: 输入包 PAWS 失败次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数:

        - ``tcp_validate_incoming()``
        - ``tcp_timewait_state_process()``
        - ``tcp_check_req()``



Listen相关
-----------
TcpExt.ListenOverflows
    :意义: 三路握手最后一步完成之后，Accept 队列超过上限的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_v4_syn_recv_sock()``


TcpExt.ListenDrops
    :意义: 任何原因导致的丢弃传入连接（SYN 包）的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_v4_syn_recv_sock()``

        包括 Accept 队列超限，创建新连接，继承端口失败等


Undo 相关
----------
TcpExt.TCPFullUndo
    :意义: TCP 窗口在不使用 slow start 完全恢复的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_ack() -> tcp_fastretrans_alert() -> tcp_try_undo_recovery()``

        Recovery 状态时，接收到到全部确认(snd_una >= high_seq)后且已经 undo 完成(undo_retrans == 0)的次数。

TcpExt.TCPPartialUndo
    :意义: TCP 窗口通过 Hoe heuristic [#]_ 机制部分恢复的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_ack() -> tcp_fastretrans_alert() -> tcp_try_undo_recovery()``

        Recovery 状态时，接收到到全部确认(snd_una >= high_seq)后且已经 undo 完成(undo_retrans == 0)的次数。

.. [#] 我也不知道这是啥，从 ``netstat`` 里看到的


TcpExt.TCPDSACKUndo
    :意义:
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_ack() -> tcp_fastretrans_alert() -> tcp_try_undo_dsack()``

        Disorder状态下，undo 完成(undo_retrans == 0)的次数。

TcpExt.TCPLossUndo
    :意义:
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_ack() -> tcp_fastretrans_alert() -> tcp_try_undo_loss()``

        Loss 状态时，接收到到全部确认(snd_una >= high_seq)后且已经 undo 完成(undo_retrans == 0)的次数。


快速路径与慢速路径
------------------
TcpExt.TCPHPHits
    :意义: 包头预测命中的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_rcv_established()``

        如果有 skb 通过“快速路径”进入到 ``sk_receive_queue`` 上，计数器加1。
        特别地，Pure ACK 以及直接复制到用户态内存上的都不算在这个计数器上。

TcpExt.TCPHPHitsToUser
    :意义: 包头预测命中并且 skb 直接复制到了用户态内存中的次数
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_rcv_established()``

TcpExt.TCPPureAcks
    :意义: 接收慢速路径中的 Pure ACK 数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_ack()``

TcpExt.TCPHPAcks
    :意义: 接收到进入快速路径的 ACK 包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无

    .. note::
        相关函数 ``tcp_ack()``

未归类
------
TcpExt.OutOfWindowIcmps
    :意义: 因为指定的数据已经在窗口外而被丢弃的 ICMP 包数量
    :取值: 0 - 无上限，整数，单调递增(COUNTER类型)
    :Tags: 无


暂时无解释
----------

暂时无解释::

    TcpExt.LockDroppedIcmps
    TcpExt.ArpFilter
    TcpExt.TCPPrequeued
    TcpExt.TCPDirectCopyFromBacklog
    TcpExt.TCPDirectCopyFromPrequeue
    TcpExt.TCPPrequeueDropped
    TcpExt.TCPRcvCollapsed
    TcpExt.TCPSACKDiscard
    TcpExt.TCPSackShifted
    TcpExt.TCPSackMerged
    TcpExt.TCPSackShiftFallback
    TcpExt.TCPBacklogDrop
    TcpExt.TCPMinTTLDrop
    TcpExt.TCPDeferAcceptDrop
    TcpExt.IPReversePathFilter
    TcpExt.TCPReqQFullDoCookies
    TcpExt.TCPReqQFullDrop
    TcpExt.TCPRcvCoalesce
    TcpExt.TCPOFOQueue
    TcpExt.TCPOFODrop
    TcpExt.TCPOFOMerge
    TcpExt.TCPChallengeACK
    TcpExt.TCPSYNChallenge
    TcpExt.TCPSpuriousRtxHostQueues
    TcpExt.BusyPollRxPackets
