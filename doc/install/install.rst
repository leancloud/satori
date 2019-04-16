初次安装
========

机器要求
--------

请找一个干净的虚拟机，内存大一点，要有 SSD，不要部署其他的服务。
LeanCloud 用了一个 2 cores, 8GB 的内存的 VM，目测这个配置可以撑住大概 2000 个左右的节点（InfluxDB 查询不多的情况下）。

交互式安装
----------

:file:`satori` 子目录中有一个 :file:`install` 脚本，以一个可以 ssh 并且 sudo
的用户执行它就可以了。
不要用 root，因为安装后会以当前用户的身份配置规则仓库。

执行下面的命令：

.. code-block:: bash

    git clone https://github.com/leancloud/satori
    cd satori/satori
    ./install

install 时会问你几个问题，填好以后会自动的生成配置，
build 出 docker image 并启动。这之后设置好DNS和防火墙就可以用了。
在仓库的第一层会有编译好的 :file:`satori-agent` 和
:file:`agent-cfg.yaml` 可以用来部署，命令行如下：

.. code-block:: bash

    /usr/bin/satori-agent -c /etc/satori-agent/agent-cfg.yaml


安装的时候会检查一下需要的组件，如果没有安装的话会进行安装。
安装后再次运行 ``./install`` 即可。

非交互式（无人工干预/批量）安装
-------------------------------

请创建一个配置文件：

.. code-block:: bash

    USE_MIRROR=1  # 或者留空，是否使用国内的镜像
    DOMAIN="www.example.com"  # 外网可以访问的域名
    INTERNAL_DOMAIN="satori01"  # 内网可以访问的域名
    RULES_REPO="/home/app/satori-conf"  # 规则仓库的地址
    RULES_REPO_SSH="app@www.example.com:/home/app/satori-conf"  # 外网可以访问的 git 仓库地址
    RESOLVERS="223.5.5.5,119.29.29.29,ipv6=off"  # 使用逗号(`,`)分隔的 DNS IP地址列表

保存为 :file:`/tmp/install.conf` 。

然后执行：

.. code-block:: bash

    git clone https://github.com/leancloud/satori
    cd satori/satori
    ./install -f /tmp/install.conf


Agent 安装
----------

如果你使用 Ansible 和 SystemD，那么在 `deploy/ansible` 中有一份样例 ansible playbook，可以参考一下。


.. note::

    agent 不会自己 daemonize，如果用了传统的 ``/etc/init.d`` 脚本的方式部署，需要注意这个问题。


登录 Web 界面
-------------

Web 界面的密码会在安装后给出，注意提示。

如果没有看到，可以在 clone 出规则仓库之后在内部找到 :file:`bitbar-plugin.py` ，
可以看到写入的 Web 界面的用户名和密码。

.. note::
    建议在测试成功后自己搭建并开启 TLS 客户端验证，或者使用 Kerberos 认证。
    这两种方式在 `/etc/satori/nginx` 中都有未开启的样例配置。
