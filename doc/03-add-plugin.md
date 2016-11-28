# 说明

这一份文档应该是放在规则仓库中的，所有的路径都是以规则仓库为根目录，也就是项目中的 `satori-rules`。
如果你是在规则仓库中看到的这个文档，这个说明可以忽略。

# 添加插件

插件分为两种，一种接受参数，另外一种不接受。

## 带参数的插件

所有接受参数的插件都在 `plugin/_metric` 目录里，命名需要跟报告的 metric 一致，
比如 `plugin/_metric/net.port.listen` 会上报 `net.port.listen` 这个 metric 。
接受参数的插件需要在在运行的时候通过 `stdin` 读取 json 格式的参数：

```json
[
  {"_metric": "net.port.listen", "_step": 30, "port": 6379},
  {"_metric": "net.port.listen", "_step": 30, "port": 3306}
]
```

参数是在规则配置中指定的，比如

```clojure
(def mysql-and-redis-rules
  (where (host "mysql-and-redis-host")
    (plugin-metric "net.port.listen" 30 {:port 3306})
    (plugin-metric "net.port.listen" 30 {:port 6379})
    (your-other-rules ...)))
```

这里 `plugin-metric` 的参数分别是 metric 名字、采集周期、插件的参数。
metric 名字和采集周期会在参数里以 `_metric` 和 `_step` 作为 key 传递进去。

之后，插件需要输出如下格式的 json :

```json
[
  {
    "endpoint": "mysql-and-redis-host",
    "tags": {"port": 3306},
    "timestamp": 1431349763,
    "metric": "net.port.listen",
    "value": 1,
    "step": 30
  },
  {
    "endpoint": "mysql-and-redis-host",
    "tags": {"port": 6379},
    "timestamp": 1431349763,
    "metric": "net.port.listen",
    "value": 0,
    "step": 30
  }
]
```

插件参数（比如上个例子中的 port）可以任意添加。
注意，这两个 json 都是 list。

|Key|Value|
|---|-----|
|endpoint|机器名，在 riemann（写报警规则时）用 `host` 表示。|
|tags|标签。比如上个例子中的端口监听，会同时监听两个端口。在 riemann 中会直接附加到 event 上。|
|timestamp|时间戳，事件发生的时间。|
|metric|监控项的名字。在 riemann 中用 `service` 表示。|
|value|监控项的值。在 riemann 中用 `metric` 表示。|
|step|采集周期，单位是秒。|

插件的输出格式基本与 open-falcon 的插件格式相同，但是 tags 是个 object，不是拼接的字符串。

插件中的 Key 跟 riemann 中对应不上是一个比较恼人的坑，需要特别注意。


## 普通插件（不带参数的）

普通插件的命名需要类似于 `30_nginx.py` 这样的，`_` 前面需要是数字，表示采集的周期。
插件输出的格式跟跟带参数插件格式一致（上面的 json）。

之后可以在规则中配置使用这个插件（假设你把这个插件放到了 `plugin/nginx` 目录）

```clojure
(def some-nginx-related-rules
  (where (host "nginx-machine")
    (plugin-dir "nginx")
    (your-other-rules ...)))
```
