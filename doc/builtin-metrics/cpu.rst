.. _builtin-cpu:

CPU 占用
========

cpu.idle
   :意义: CPU 空闲时间百分比
   :取值: 0 - 100，与 CPU 核数无关
   :Tags: 无

cpu.busy
    :意义: CPU 忙时间百分比
    :取值: 0 - 100，与 CPU 核数无关
    :Tags: 无

    .. note::
        与 ``cpu.idle`` 的和总是100

cpu.user
    :意义: 用来运行用户态代码的 CPU 时间百分比
    :取值: 0 - 100，与 CPU 核数无关
    :Tags: 无

    .. note::
        通常 ``cpu.user`` 占用意味着正在运行业务逻辑

cpu.nice
    :意义: 用来运行低优先级进程的用户态代码的 CPU 时间百分比
    :取值: 0 - 100，与 CPU 核数无关
    :Tags: 无

    .. note::
        可以通过 ``nice`` 命令启动低优先级进程，或者 ``top`` 中的 renice 功能调整优先级

        低优先级进程会被正常进程抢占，所以考虑 CPU 占用时可以当做是空闲

cpu.system
    :意义: 用来运行内核态代码的 CPU 时间百分比
    :取值: 0 - 100，与 CPU 核数无关
    :Tags: 无

    .. note::
        内核态代码通常在进行系统调用、中断、换页的时候执行

cpu.iowait
    :意义: 用来等待 IO 的 CPU 时间百分比
    :取值: 0 - 100，与 CPU 核数无关
    :Tags: 无

    .. note::
        IO wait 不意味着 CPU 被浪费掉了，只是现在只能等待做 IO 的进程，此时如果有需要 CPU 时间的进程还是会被调度的，跟 idle 有点像

cpu.irq
    :意义: 用来处理中断的 CPU 时间百分比
    :取值: 0 - 100，与 CPU 核数无关
    :Tags: 无

    .. note::
        这里是指不能被抢占的中断处理，实际上大部分的中断处理的部分会放在软中断中处理，软中断可以被抢占

cpu.softirq
    :意义: 用来处理软中断 CPU 时间百分比
    :取值: 0 - 100，与 CPU 核数无关
    :Tags: 无

    .. note::
        软中断通常是中断的『bottom half』，是实际的中断处理部分，比如网卡在接受数据包并传给网络栈的 CPU 会被计算在软中断里（也是最常见的软中断占满的原因）

cpu.steal
    :意义: 被宿主机截取的 CPU 时间百分比
    :取值: 0 - 100，与 CPU 核数无关
    :Tags: 无

    .. note::
        半虚拟化环境中存在，宿主机会根据情况不再调度当前虚拟机，此时的 CPU 时间会被计算成 steal

        并不是每个 VM 环境都会看到这个 metric，有一些环境会把 steal 算在 system 里

cpu.guest
    :意义: 用来执行虚拟机代码的 CPU 时间百分比
    :取值: 0 - 100，与 CPU 核数无关
    :Tags: 无

    .. note::
        虚拟化环境的宿主机中存在，当 CPU 在执行 VM 内的代码时算在 guest 里

cpu.switches
    :意义: 进程上下文切换数
    :取值: 0 - 无上界，整数
    :Tags: 无

    .. note::
        单调递增（COUNTER 类型）

        当一个 CPU 核由运行着一个进程切换到运行另一个进程时，``cpu.switches`` 加1
