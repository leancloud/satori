.. _repo-signing:

规则仓库签名
============

为了防止监控机被入侵导致可以随意向被监控机推送任意的可执行文件（插件），
Satori 提供了可选的签名机制。

在安装初次安装完成后，签名机制是没有打开的。

生成签名用的 key
-----------------

在安装完成后，可以先 clone 下来规则仓库，并且在规则仓库中找到 ``sign`` 工具。

通过如下命令生成密钥对：

.. code-block:: bash

   ./sign --generate


执行这个命令之后，私钥会被保存在 :file:`~/.satori-signkey` 中，公钥会直接告诉你。

如果你弄丢了公钥，可以通过如下命令获得公钥：

.. code-block:: bash

   ./sign --get-public

这个命令会根据你的私钥重新计算出公钥来。

.. warning::
    私钥丢失无法恢复。


对规则仓库进行签名
------------------

在修改过规则仓库并且 commit 之后，执行如下命令

.. code-block:: bash

    ./sign

就会以 :file:`~/.satori-signkey` 作为私钥对当前 commit 进行签名了，签名后就可以 push 到监控机的规则仓库中并开始生效了。

签名是通过 ``git amend`` 修改最后一个 commit 的 commit message 来嵌入的，可以通过 ``git log`` 命令观察到。

.. warning::
   如果不小心重新签名了已经 push 出去的 commit，可以通过 ``git reflog``
   找到之前的 commit 并 ``git reset --hard`` 过去。


配置 agent 接受的签名公钥
-------------------------

在 :ref:`agent-config-sample` 能看到 ``signingKeys`` 的配置，将你的公钥填进去就好。

.. code-block:: yaml

    plugin:
      signingKeys:
        - owner: You
          key: NotAVaildKeyGetOneKeyBySignToolInRulesRepo==
        - owner: youmu
          key: RYzIU5vEK3e4asrN0KrPpNdvjBRQq+3Mva5z27ba9sw=

其中 ``key`` 是你的 **公钥** ，其他的都是元信息，可以任意添加删除。

agent 按照配置了 ``signingKey`` 的新配置启动后，就会在每次更新规则仓库后，先验证签名，再检出最新的 commit。


另外，``authorizedKeys`` 选项指定了一个规则仓库中的文件名（默认为 ``authorized_keys.yaml`` ），
这个文件中配置的额外的 key 也都会被 agent 接受，
但是修改了这个配置的 commit 必须由 ``signingKeys`` 指定的 key 来签名。

这个机制用来方便的增删签名 key 而不需要修改所有机器上 agent 的配置。


阻止未签名的 push
-----------------

开启了签名机制后会经常忘了先签名再 push，所以在 :file:`deploy/rules-repo-pre-receive-hook`
提供了一个 git hook，将这个文件复制到 :file:`/path/to/rules-repo-on-server/.git/hooks/pre-receive`
并且加上可执行权限，git 就会阻挡没有签名的 push。

.. warning::
   这个 hook 只会验证签名是否存在，不会验证有效性，请不要依赖这个机制作为安全措施！
