# 介绍

Satori 是一个由 LeanCloud 维护的监控系统。

起初在 LeanCloud 内部是直接使用了 [Open-Falcon](http://open-falcon.org/) 。
后续的使用过程中因为自己的需求开始做修改，形成的现在的这样的结构。

# 截图
![Nodes](https://raw.githubusercontent.com/leancloud/satori/master/doc/images/nodes.png)
![Events](https://raw.githubusercontent.com/leancloud/satori/master/doc/images/events.png)
![BitBar](https://raw.githubusercontent.com/leancloud/satori/master/doc/images/bitbar.png)

# 文档
完整文档麻烦看[这里](http://satori-monitoring.rtfd.io)。

# 交流
如果你有什么建议，欢迎在 issue 里面说一下。
也有 QQ 群可以参与讨论：554765935

# 怎么安装？

请找一个干净的虚拟机，内存大一点，要有 SSD，不要部署其他的服务。
LeanCloud 用了一个 2 cores, 8GB 的内存的 VM，目测这个配置可以撑住大概 2000 个左右的节点（InfluxDB 查询不多的情况下）。

执行下面的命令：

```bash
$ git clone https://github.com/leancloud/satori
$ cd satori/satori
$ ./install
```
satori 子目录中有一个 install 脚本，以一个可以 ssh 并且 sudo 的用户执行它就可以了。
不要用 root，因为安装后会以当前用户的身份配置规则仓库。

install 时会问你几个问题，填好以后会自动的生成配置，build 出 docker image 并启动。这之后设置好DNS和防火墙就可以用了。
在仓库的第一层会有编译好的 `satori-agent` 和 `agent-cfg.json` 可以用来部署，命令行如下：

`/usr/bin/satori-agent -c /etc/satori-agent/agent-cfg.json`

agent 不会自己 daemonize，如果用了传统的 `/etc/init.d` 脚本的方式部署，需要注意这个问题。

如果需要无人工干预安装，请创建一个配置文件：

```bash
USE_MIRROR=1  # 或者留空，是否使用国内的镜像
DOMAIN="www.example.com"  # 外网可以访问的域名
INTERNAL_DOMAIN="satori01"  # 内网可以访问的域名
RULES_REPO="/home/app/satori-conf"  # 规则仓库的地址
RULES_REPO_SSH="app@www.example.com:/home/app/satori-conf"  # 外网可以访问的 git 仓库地址
```

保存为 `/tmp/install.conf`。

然后执行：
```bash
$ git clone https://github.com/leancloud/satori
$ cd satori/satori
$ ./install -f /tmp/install.conf
```

# 设计思路
- Satori 希望最大程度的减少监控系统的部署维护难度。如果在任何的部署、增删维护报警的时候觉得好麻烦，那么这是个 bug。
- 监控时的需求很多样，Satori 希望做到『让简单的事情简单，让复杂的事情可能』。常用的监控项都会有模板，可以用 Copy & Paste 解决。略复杂的监控需求可以阅读 [riemann 的文档](http://riemann.io/)，来得知怎么编写复杂的监控规则。
- 因为 LeanCloud 的机器规模不大，另外再加上现在机器的性能也足够强劲了，所以放弃了 Open-Falcon 中横向扩展目标。如果你的机器数量或者指标数目确实很大，可以将 transfer、 InfluxDB、 riemann 分开部署。这样的结构支撑 5k 左右的节点应该是没问题的。
- 在考察了原有的 Open-Falcon 的架构后，发现实质上高可用只有 transfer 实现了，graph、judge 任何一个实例挂掉都会影响可用性。对于 graph，如果一个实例挂掉的话，还必须要将那个节点恢复，不能通过新加节点改配置的方式来恢复；judge 尽管可以起新节点，但是还是要依赖外部的工具来实现 failover，否则是需要修改 transfer 的配置的。因此 Satori 中坚持用单点的方式来部署，然后通过配合公网提供的 APM 服务保证可用性。真的希望有高可用的话，riemann 和 alarm 可以部署两份，通过 keepalived 的方式来实现，InfluxDB 可以官方的 Relay 来实现。

# 1 minute taste

## 简单需求（通用模板，Copy & Paste 改参数可以实现）

> 完整版请看 `satori-rules/rules/infra/mongodb.clj`

```clojure
(def infra-mongodb-rules
  ; 在 mongo1 mongo2 mongo3 ... 上做监控
  (where (host #"^mongo\d+")

    ; 执行 mongodb 目录里的插件（里面有收集 mongodb 指标的脚本）
    (plugin-dir "mongodb")

    ; 每 30s 收集 27018 端口是不是在监听
    (plugin "net.port.listen" 30 {:port 27018})

    ; 过滤一下 mongodb 可用连接数的 metric（上面插件收集的）
    (where (service "mongodb.connections.available")
      ; 按照 host（插件中是 endpoint）分离监控项，并分别判定
      (by :host
        ; 报警在监控值 < 10000 时触发，在 > 50000 时恢复
        (set-state-gapped (< 10000) (> 50000)
          ; 600 秒内只报告一次
          (should-alarm-every 600
            ; 报告的标题、级别（影响报警方式）、报告给 operation 组和 api 组
            (! {:note "mongod 可用连接数 < 10000 ！"
                :level 1
                :groups [:operation :api]})))))

    ; 另一个监控项
    (where (service "mongodb.globalLock.currentQueue.total")
      (by :host
        (set-state-gapped (> 250) (< 50)
          (should-alarm-every 600
            (! {:note "MongoDB 队列长度 > 250"
                :level 1
                :groups [:operation :api]})))))))
```

```bash
cd /path/to/rules/repo  # 规则是放在 git 仓库中的
git commit -a -m 'Add mongodb rules'
git push  # 然后就生效了
```

## 复杂需求

这是一个监控队列堆积的规则。
队列做过 sharding，分布在多个机器上。
但是有好几个数据中心，要分别报告每个数据中心队列情况。
堆积的定义是：在一定时间内，队列非空，而且队列元素数量没有下降。

> 提示：这是一个简化了的版本，完整版可以看 `satori-rules/rules/infra/kestrel.clj`

```clojure
(def infra-kestrel-rules
 ; 在所有的队列机器上做监控
 (where (host #"kestrel\d+$")
  ; 执行队列相关的监控脚本（插件）
  (plugin-dir "kestrel")

  ; 过滤『队列项目数量』的 metric
  (where (service "kestrel_queue.items")
   ; 按照队列名和数据中心分离监控项，并分别判定
   (by [:queue :region]
    ; 将传递下来的监控项暂存 60 秒，然后打包（变成监控项的数组）再向下传递
    (fixed-time-window 60
     ; 将打包传递下来的监控项做聚集：将所有的 metric 值都加起来。
     ; 因为队列监控的插件是每 60 秒报告一次，并且之前已经按照队列名和数据中心分成子流了，
     ; 所以这里将所有 metric 都加起来以后，获得的是单个数据中心单个队列的项目数量。
     ; 聚集后会由监控项数组变成单个的监控项。
     (aggregate +
      ; 将传递下来的聚集后的监控项放到滑动窗口里，打包向下传递。
      ; 这样传递下去的，就是一个过去 600 秒单个数据中心单个队列的项目数量的监控项数组。
      (moving-event-window 10
       ; 如果已经集满了 10 个，而且这 10 个监控项中都不为 0 （队列非空）
       (where (and (>= (count event) 10)
                   (every? pos? (map :metric event)))
        ; 再次做聚集：比较一下是不是全部 10 个数量都是先来的小于等于后来的（是不是堆积）
        (aggregate <=
         ; 如果结果是 true，那么认为现在是有问题的
         (set-state (= true)
          ; 每 1800 秒告警一次
          (should-alarm-every 1800
           ; 这里 outstanding-tags 是用来区分报警的，
           ; 即如果 region 的值不一样，那么就会被当做不同的报警
           (! {:note #(str "队列 " (:queue %) " 正在堆积！")
               :level 2
               :outstanding-tags [:region]
               :groups [:operation]}))))))))))))
```

# 架构
![Architecture](https://raw.githubusercontent.com/leancloud/satori/master/doc/images/arch.png)

# 与 Open-Falcon 的区别

## agent
- 支持按照正则表达式排除 metric（用例：排除 docker 引入的奇怪的挂载，netns 什么的）
- 支持从 agent 上附加固定的 tag（用例：region=cn-west）
- 支持自主的插件更新
- 支持带参数的插件
- 支持持续执行(long running)的插件
- 去除了 push 和 plugin 以外的所有 http 接口
- 去掉了单机部署的功能，现在强制要求指定一个 transfer 组件
- 不兼容 open-falcon 的 heartbeat
- 因为修改了 metric 的数据结构，与 open-falcon 的 transfer 不兼容

## transfer：
- 支持发送到 influxdb（使用了 hitripod 的补丁）
- 支持发送到 riemann
- 支持发送到 transfer（gateway 功能集成）
- 不再支持发送到 graph 和 judge
- 重构了发送端的代码，现在代码比之前容易维护了

## alarm：
- 弃用了 open-falcon 的 alarm，完全重写
- 不支持报警合并
- 支持 EVENT 类型（只报警不记录状态）
- 支持多种报警后端（电话、短信、BearyChat、OneAlert、PagerDuty、邮件、微信企业号），并且易于扩展
- Mac 下有好用的 BitBar 插件

## sender
集成进了 alarm 中。

## links
在 Satori 中移除了。推荐直接使用低优先级的通道（如 BearyChat/其他IM，或者 BitBar），不做报警合并。

## graph & query & dashboard
被 Grafana 和 InfluxDB 代替

## judge：
- 被 riemann 代替
- riemann 较 judge 相比，可以节省 60% 以上的内存，CPU占用要低 50%。

## task
在 Satori 中移除了。InfluxDB 自带 task 的功能。

## aggregator
在 Satori 中移除了。riemann 中可以轻松的实现 aggregator 的功能。

## nodata
在 Satori 中移除了。可以参见规则中的 feed-dog 和 watchdog，实现了相同的功能。

## portal
在 Satori 中移除了。报警规则通过 git 仓库管理。

## gateway:
合并进了 transfer

## hbs
- 在 Satori 中叫做 master
- 与 hbs 不兼容
- 不再将节点数据记录到数据库中，没有数据库的依赖。

## uic
Satori 中去除，直接在规则仓库中编辑用户信息。

## fe
完全重写，采用了纯前端的方案（frontend 文件夹）。
