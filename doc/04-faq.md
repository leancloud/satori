# 常见的问题

## Satori 更新了，我怎么升级？

```bash
$ cd satori/satori
$ ./install rebuild
$ sudo docker-compose up -d
```

## 重启了 master，机器信息都丢掉了？
master 不会持久化保存机器信息，只会在内存中记录发来心跳的 agent 的信息。 这种情况等两分钟就可以了。

如果长时间都是空白，可以检查一下 agent 的日志。

agent 的日志会输出到 stdout、stderr 上，所以日志文件具体在哪里就跟你的部署方式有关了，比如，upstart 的日志可以在 `/var/log/upstart` 中找到。

## 机器信息里看不到插件项，插件版本也不对？
这种情况是因为 master 没有收到来自监控规则的配置，通常在重启了 master 之后会出现这种情况。

触发一下规则更新就可以了，可以重启下 riemann 中的 reloader（重启 riemann 的 docker 容器也可以），或者做一个小的规则修改，push 上去，两分钟后就会恢复。

这种情况下，只要 agent 不重启，之前的配置就仍然有效。

## 我修改了规则之后，为什么相关报警一直留在报警页面里？
报警页面的记录是由 riemann 发给 alarm 的，恢复时也是。

修改了规则之后，riemann 就不会发送恢复的事件，alarm 会一直保留这个事件。

可以观察下触发时间，如果时间很长的话，就是这种情况了。

直接删除这个事件就可以了。


## 插件目录中的插件总是不执行？
插件需要 `chmod +x` ，以及正确的 shebang (#!/bin/bash 这个东西)。
