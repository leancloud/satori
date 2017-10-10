.. _alarm-config:

报警配置
========

Satori 中的报警是由 alarm 组件负责的，alarm 组件的配置都在规则仓库中的 :file:`alarm` 目录中。

:file:`alarm` 目录中的所有 yaml 文件都会解析合并，所以可以任意的拆分文件。

LeanCloud 内部的报警级别是这么设定的：

+----+-----------------------------------------+
| LV | 效果                                    |
+====+=========================================+
| 0  | 语音通话                                |
+----+-----------------------------------------+
| 1  | 短信（包括运维手机)                     |
+----+-----------------------------------------+
| 2  | 短信                                    |
+----+-----------------------------------------+
| 3  | 微信企业号                              |
+----+-----------------------------------------+
| 4  | 邮件（但是0123级别的不会发）            |
+----+-----------------------------------------+
| 5  | BearyChat                               |
+----+-----------------------------------------+
| 6  | 什么都不做（但是会显示在 BitBar 插件中) |
+----+-----------------------------------------+

规则仓库的默认配置就是按照这个来的，可以自己按照需求修改。


.. _alarm-config-users-and-groups:

报警人员和组配置
----------------

样例配置如下，感觉不用解释了……

其中 ``users`` 下每一个人员的配置需要填 ``name`` 表示名字，以及一个可选的 ``threshold`` 表示『过滤级别在这之下的报警』。 剩余的字段是给报警的 backend 的参数，需要参考各个 backend 的文档。


.. code-block:: yaml

    groups:
      operation:
        - proton
        - phone_ops
        - bearychat_ops
        - pagerduty_ops

    users:
      phone_ops:
        name: "运维手机"
        phone: 18888888888
        threshold: 1

      proton:
        name: "Proton"
        email: proton@example.com
        phone: 1888888888
        wechat: proton

      bearychat_ops:
        name: "运维 BearyChat 机器人"
        bearychat: "[PUT_YOUR_URL_HERE]"

      pagerduty_ops:
        name: "运维 PagerDuty"
        pagerduty: "PUT_YOUR_KEY_HERE"
        onealert: "PUT_YOUR_KEY_HERE"

报警策略配置
------------

在 ``strategies`` 下的每一个策略需要提供至少 2 个参数：

+---------+-------------------------------------------------------+
| Key     | 意义                                                  |
+=========+=======================================================+
| backend | 策略的 backend [#]_                                   |
+---------+-------------------------------------------------------+
| level   | 当前策略处理的报警级别，是个 string，可以指定多个级别 |
+---------+-------------------------------------------------------+


.. [#] backend 就是在 Satori 源码的 :file:`alarm/src/backend` 中的用于发送报警的小段代码，
       可以自己扩展。

剩下的参数是提供给 backend 用的。

.. code-block:: yaml

    strategies:
      phonecall:
        backend: nexmo_tts
        level: '0'
        api_key: 'foooo'
        api_secret: 'barrrr'
        voice: female
        lg: zh-cn
        repeat: 3
        prefix: ''

      sms:
        backend: yunpian_sms
        level: '012'
        signature: 'LC报警'
        api_key: '812912897398172387893687401298'


SMTP 发送邮件
-------------

Backend 名字:
    smtp

报警策略中需要配置的参数:
    +-----------+--------------------+
    | Key       | 意义               |
    +===========+====================+
    | server    | 邮件服务器地址     |
    +-----------+--------------------+
    | send_from | 邮件的 Sender 地址 |
    +-----------+--------------------+
    | username  | SMTP 认证用户名    |
    +-----------+--------------------+
    | password  | SMTP 认证密码      |
    +-----------+--------------------+

用户中需要配置的参数:
    +-------+--------------------+
    | Key   | 意义               |
    +=======+====================+
    | email | 当前用户的邮件地址 |
    +-------+--------------------+

配置样例:
    .. code-block:: yaml

        strategies:
          email:
            backend: smtp
            level: '4'
            server: smtp.mailgun.org
            send_from: satori-alarm@example.com
            username: fooooo
            password: barrrr

        users:
          example:
            name: "例子"
            email: example@example.com


发送短信
--------

这里短信使用了 `云片 <https://www.yunpian.com>`__ 的服务，需要在上面注册账号，获得 API key

Backend 名字:
    yunpian_sms

报警策略中需要配置的参数:
    +-----------+---------------+
    | Key       | 意义          |
    +===========+===============+
    | api_key   | API Key       |
    +-----------+---------------+
    | signature | 短信签名 [#]_ |
    +-----------+---------------+

    .. [#] 就是运营商强制在【】中加入的信息

用户中需要配置的参数:
    +-------+----------------+
    | Key   | 意义           |
    +=======+================+
    | phone | 当前用户的手机 |
    +-------+----------------+

配置样例:
    .. code-block:: yaml

        strategies:
          sms:
            backend: yunpian_sms
            level: '012'
            signature: '报警'
            api_key: PUT_YOUR_KEY_HERE

        users:
          example:
            name: "例子"
            phone: 18888888888

电话报警
--------

这里使用了 `Nexmo <https://www.nexmo.com>`__ 的服务，需要在上面注册账号，获得 API key

Backend 名字:
    nexmo_tts

报警策略中需要配置的参数:
    +------------+-----------------------------+
    | Key        | 意义                        |
    +============+=============================+
    | api_key    | API Key                     |
    +------------+-----------------------------+
    | api_secret | API Secret                  |
    +------------+-----------------------------+
    | voice      | 语音声音，可以填 ``female`` |
    +------------+-----------------------------+
    | lg         | 语言， ``zh-cn`` 为中文     |
    +------------+-----------------------------+
    | repeat     | 重复次数                    |
    +------------+-----------------------------+
    | prefix     | 在报警标题前加的固定的话    |
    +------------+-----------------------------+

用户中需要配置的参数:
    +-------+----------------+
    | Key   | 意义           |
    +=======+================+
    | phone | 当前用户的手机 |
    +-------+----------------+

配置样例:
    .. code-block:: yaml

        strategies:
          phonecall:
            backend: nexmo_tts
            level: '0'
            api_key: PUT_YOUR_KEY_HERE
            api_secret: PUT_YOUR_KEY_HERE
            voice: female
            lg: zh-cn
            repeat: 3
            prefix: ''

        users:
          example:
            name: "例子"
            phone: 18888888888


微信企业号
----------

.. warning::
    腾讯现在只提供企业微信了，所以不再提供文档


BearyChat
---------

这个会 POST 到 BearyChat 的 Incoming 机器人上。

Backend 名字:
    bearychat

报警策略中需要配置的参数:
    无

用户中需要配置的参数:
    +-----------+--------------------------------+
    | Key       | 意义                           |
    +===========+================================+
    | bearychat | 当前用户的 Incoming 机器人地址 |
    +-----------+--------------------------------+

配置样例:
    .. code-block:: yaml

        strategies:
          bearychat:
            backend: bearychat
            level: '012345'

        users:
          example:
            name: "例子"
            bearychat: "https://hook.bearychat.com/=foobar/incoming/bazbazbazbazbazabaz"

PagerDuty
---------

Backend 名字:
    pagerduty

报警策略中需要配置的参数:
    无

用户中需要配置的参数:
    +-----------+----------------------------------+
    | Key       | 意义                             |
    +===========+==================================+
    | pagerduty | 当前用户的 PagerDuty service_key |
    +-----------+----------------------------------+

配置样例:
    .. code-block:: yaml

        strategies:
          pagerduty:
            backend: pagerduty
            level: '012345'

        users:
          example:
            name: "例子"
            pagerduty: "abcdefg123123123123123"

OneAlert
--------

Backend 名字:
    onealert

报警策略中需要配置的参数:
    无

用户中需要配置的参数:
    +----------+-----------------------------+
    | Key      | 意义                        |
    +==========+=============================+
    | onealert | 当前用户的 OneAlert app key |
    +----------+-----------------------------+

配置样例:
    .. code-block:: yaml

        strategies:
          onealert:
            backend: onealert
            level: '012345'

        users:
          example:
            name: "例子"
            onealert: "abcdefg123123123123123"


静默（不报警）
--------------

Backend 名字:
    noop

报警策略中需要配置的参数:
    无

用户中需要配置的参数:
    无

配置样例:
    .. code-block:: yaml

        strategies:
          indicator:
            backend: noop
            level: '0123456'

        users:
          example:
            name: "例子"

.. note::
    通常用作最低优先级的报警。静默的报警会出现在 Web UI 和 BitBar Plugin 中。

    BitBar Plugin 的插件配置可以参考 Web UI 首屏的说明。
