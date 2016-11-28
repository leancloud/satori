# 说明

这一份文档应该是放在规则仓库中的，所有的路径都是以规则仓库为根目录，也就是项目中的 `satori-rules`。
如果你是在规则仓库中看到的这个文档，这个说明可以忽略。

# 添加规则

规则需要放在 `rules` 文件夹的一个子文件夹中，推荐直接复制一个已有的文件在这之上修改。
规则中的以 `-rules` 结尾的函数（或者 riemann 的流，其实都一样）会自动添加到 riemann 中。

比如，我想添加一个 `foo/bar.clj` 作为 foo 服务下关于 bar 的监控，
那么就随便找一个文件夹中的规则（最外层的不行！），然后修改 ns 和规则的名字
（推荐 Copy & Paste！代码很容易看懂但是写起来可能很麻烦）

```clojure
(ns infra.memcache ; 这里要修改
  (:use riemann.streams
        agent-plugin
        alarm))

(def infra-memcache-rules ; 这里名字要修改
  (where (host #"cache\d+$") ; 规则当然也要改
    (plugin-dir "memcache")
    (plugin-metric "net.port.listen" 30 {:port 11211})

    (where (and (service "net.port.listen")
                #(= (:port %) 11211))
      (by :host
        (set-state-gapped (< 1) (> 0)
          (should-alarm-every 120
            (! {:note "memcache 端口不监听了！"
                :level 0
                :groups [:operation :api]})))))))
```

修改成

```clojure
(ns foo.bar
  (:use riemann.streams
        agent-plugin
        alarm))

(def foo-bar-rules ; 随便叫什么但是一定要以 -rules 结尾
  (where (host "host-which-run-foo-bar")
    (plugin-dir "foo/bar") ; 要运行 foo/bar 目录里的插件
    (plugin-metric "net.port.listen" 30 {:port 12345}) ; 运行 net.port.listen 插件，30秒一次，以及参数

    (where (and (service "net.port.listen")
                #(= (:port %) 12345)) ; 匹配 net.port.listen 插件收集的 metric
      (by :host ; 按照 host 分成子流（这个例子里就只有一个 host 所以并不会分）
        (set-state-gapped (< 1) (> 0)  ; 根据条件设置事件的 :state
          (should-alarm-every 120  ; 如果是 :problem 每120秒发一次报警。
            (! {:note "foobar 服务端口不监听了！"
                :level 0
                :groups [:operation :api]})))))))
```

之后 commit 然后 push 上去就会生效了。


# 错误处理
可以在 `riemann.config` 中找到相关的规则，可以修改报警级别和报警组。
目前 LeanCloud 的做法是将这些信息发送到 BearyChat。
监控规则 Reload 失败后还会用老规则，所以不用担心。

# 关于报警
流向 `!` 这个流的事件都会报警。
但是不能直接把事件喂给 `!`，因为 `!` 并不知道这个事件是正常还是有问题的状态，
所以需要指定事件的状态是 `:ok` 还是 `:problem`。
通常可以用 `set-state` 和 `set-state-gapped` 这两个流来完成。

最简单的一个例子：

```clojure
(def app-important-rules
  (where (service "service.very.important.latency")
    (set-state-gapped (> 1000) (< 100)  ; 当 latency 超过 1000 后报警，回落到 100 以下变成正常状态
        (! {:note "报警标题，标题对于一个特定的报警是不能变的（不要把报警的数据编码在这里面）"
            :level 1  ;报警级别, 0最高，6最小。报警级别影响报警方式。
            :event? false  ; 可选，是不是事件（而不是状态）。默认 false。如果是事件的话，只会发报警，不会记录状态（alarm插件里看不到）。
            :expected 233  ; 可选，期望值，暂时没用到
            :outstanding-tags [:region :mount]  ; 可选，相关的tag，写在这里的 tag 会用于区分不同的事件，以及显示在报警内容中, 不填的话默认是所有的tag
            :groups [:operation]})  ; groups 是在规则仓库的 alarm 配置里管理的)
```

报警级别是在 alarm 的配置中定义的，具体可以看一下 `02-config-alarm.md` 文件。
