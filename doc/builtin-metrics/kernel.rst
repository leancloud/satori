.. _builtin-kernel:

内核参数
========

kernel.maxfiles
   :意义: 整个系统可以打开的 fd 数量
   :取值: 0 - 无上限，整数
   :Tags: 无

kernel.files.allocated
   :意义: 整个系统已经打开的 fd 数量
   :取值: 0 - ``kernel.maxfiles`` ，整数
   :Tags: 无

kernel.files.left
   :意义: 整个系统剩余可分配的 fd 数量
   :取值: 0 - ``kernel.maxfiles`` ，整数
   :Tags: 无

kernel.maxproc
   :意义: 整个系统可以创建的进程数量
   :取值: 0 - 无上限，整数
   :Tags: 无

   .. note::
      实际上指的是 Task 数量，Linux 并不会严格的区分进程/线程，
      而是统一当做 Task 来调度。

      这个值也可以同时是最大的 pid 值。

