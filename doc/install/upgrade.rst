更新
====

因为规则仓库是开放修改的，与更新同步会比较麻烦，需要手工操作。
所以这里会记录更新相关的步骤，怎么更新内置规则、配置文件等等。

没有大变化的话该怎么更新？
--------------------------

.. code-block:: bash

    cd satori/satori
    git pull
    ./install rebuild
    sudo docker-compose up -d
