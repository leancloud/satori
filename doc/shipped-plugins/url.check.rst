.. _url-check:

URL 监控
========

收集满足过滤条件的进程数量

插件文件地址
    url.check

插件类型
    接受参数，重复执行


插件参数
--------

+---------+-----------------------------------------------------+
| 参数    | 功能                                                |
+=========+=====================================================+
| url     | 需要监控的 URL                                      |
+---------+-----------------------------------------------------+
| method  | **可选** ，HTTP method，默认是 `GET`                |
+---------+-----------------------------------------------------+
| params  | **可选** ，URL Query String                         |
+---------+-----------------------------------------------------+
| headers | **可选** ，额外的 HTTP 头                           |
+---------+-----------------------------------------------------+
| data    | **可选** ，POST 数据                                |
+---------+-----------------------------------------------------+
| match   | **可选** ，对返回结果做匹配                         |
+---------+-----------------------------------------------------+
| timeout | **可选** ，访问超时时间，默认是 5 秒                |
+---------+-----------------------------------------------------+
| verify  | **可选** ，若 URL 是 https 链接，是否验证证书合法性 |
+---------+-----------------------------------------------------+
| name    | **可选** ，这个监控的名字，用来区分其他的监控       |
+---------+-----------------------------------------------------+

.. note::
   ``params``、 ``headers``、 ``match`` 都是 Key-Value 对（Python 的 dict，或者 clojure 里的 assoc）。

   ``match`` 中的 key 代表了 match 的名字，值代表了需要 match 的内容的正则表达式。


上报的监控值
------------

url.check.time
    :意义: 访问指定 URL 的请求时间
    :取值: 0 - 无上限，浮点数，单位：秒
    :Tags: {"name": " ``指定的监控名字`` "}

url.check.status
    :意义: 访问指定 URL 的 HTTP Status Code
    :取值: 整数
    :Tags: {"name": " ``指定的监控名字`` "}

    .. note::
        按照 Cloudflare 的约定新增了两个 Code:

        +------+---------------------------------------------------------+
        | Code | 意义                                                    |
        +======+=========================================================+
        | 524  | 成功建立了 TCP 连接并发送了请求，但是服务器没有及时响应 |
        +------+---------------------------------------------------------+
        | 521  | 无法建立 TCP 连接                                       |
        +------+---------------------------------------------------------+

url.check.content-length
    :意义: 访问指定 URL 返回的内容的长度
    :取值: 0 - 无上限，整数，单位：Byte
    :Tags: {"name": " ``指定的监控名字`` "}

url.check.match. ``key``
    :意义: ``match`` 中 ``key`` 指定的正则表达式在返回的内容中出现的次数
    :取值: 0 - 无上限，整数
    :Tags: {"name": " ``指定的监控名字`` "}


监控规则样例
------------

.. code-block:: clojure

    (def check-baidu-rules
      (where (host #"^host1")
        (plugin "url.check" 30
          {:name "check-baidu",
           :method "GET"
           :params {:foo "bar"}
           :headers {:User-Agent "Monitoring script from Satori"}
           :timeout 3
           :verify true
           :match {:html5 "<!DOCTYPE html>",
                   :not-exist "I'll be surprised if this strings exists!"}
           :url "https://www.baidu.com"})

        (where (and (service "url.check.status")
                    (= (:name event) "check-baidu"))
          (by :host
            (adjust [:metric int]
              (judge (not= 200)
                (runs 3 :state
                  (should-alarm-every 300
                    (! {:note "百度挂了！"
                        :level 3
                        :expected true
                        :groups [:operation]})))))))

        (where (and (service "url.check.time")
                    (= (:name event) "check-baidu"))
          (by :host
            (judge (> 1)
              (runs 3 :state
                (should-alarm-every 300
                  (! {:note "百度好卡！"
                      :level 3
                      :expected true
                      :groups [:operation]})))))))

        (where (and (service "url.check.match.not-exist")
                    (= (:name event) "check-baidu"))
          (by :host
            (judge (> 0)
              (runs 3 :state
                (should-alarm-every 300
                  (! {:note "百度被我们入侵了咩哈哈！"
                      :level 3
                      :expected true
                      :groups [:operation]})))))))
