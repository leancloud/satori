更新
====

因为规则仓库是开放修改的，与更新同步会比较麻烦，需要手工操作。
所以这里会记录更新相关的步骤，怎么更新内置规则、配置文件等等。


从 1.x 更新到 2.x
-----------------

1. 首先参照下面的 :ref:`update-trivial` 更新
2. 将规则仓库 ``alarm`` 下的 ``backends`` 和 ``hooks`` 复制到你的规则仓库的相应位置
3. 将你自己的规则仓库中的 ``_metric`` 下的所有插件移动到上层目录，与其他插件目录处于同一级
4. 将规则仓库 ``rules`` 下最新的 ``agent_plugin.clj`` 、 ``alarm.clj`` 、 ``lib.clj`` 、 ``once.clj`` 、 ``riemann.config`` 复制到你的规则仓库内的相应位置
5. 修改所有的监控规则文件，确保最上面的 ``(ns ...)`` 语句引用到了 ``agent_plugin`` ``alarm`` 和 ``lib`` 的所有符号（参见新规则仓库中的规则）
6. 提交、签名（如果开启）并 push 你的规则仓库，并且到机器上的规则仓库执行 ``git reset --hard`` 以确保 Satori 能读到最新的规则
7. 通过 ``sudo docker-compose restart`` 重新启动所有服务
8. 通过 ``sudo docker-compose logs <XXX>`` 命令查看 ``riemann`` 和 ``alarm`` 的日志，确认服务可以正常工作。 如果 ``riemann`` 不能正常工作，请按照报错提示确认是否遗漏了上述的升级过程
9. 组件都能正常工作后，更新 agent 到最新版


.. _update-trivial:

没有大变化的话该怎么更新？
--------------------------

.. code-block:: bash

    cd satori/satori
    git pull
    ./install rebuild
    sudo docker-compose up -d
