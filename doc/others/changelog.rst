.. _changelog:

ChangeLog
---------

2.0.0
=====

相对于 1.x 更新了以下东西：

1. alarm 的后端放进了规则仓库，增加报警方式不再需要 fork alarm
2. alarm 增加了 hook 机制，用来在发送报警的时候对报警做变换，或者阻止发送
3. 报警规则现在可以附带元信息
4. riemann 去掉了经常出问题的 reloader，将重载规则的逻辑直接集成在规则内
5. nginx 改为 openresty 并集成了自动申请 Let's Encrypt 证书的功能，不用再为了 SSL 头疼了
6. InfluxDB Riemann Grafana 都升级成了最新版
7. NVIDIA GPU 收集插件
8. 集群交叉探测插件（Ping）重写
9. 遇到 agent 自己的 crash 现在也能通过报警机制发送出来了
10. agent 现在可以选择使用 FQDN 作为自己的机器名（可配置）
11. 添加了 Kerberos(SPNEGO) 验证的支持

不兼容的改变：
1. agent 不再将接受参数的插件放在单独的 ``_metric`` 目录中
2. agent 内置的内存指标 ``mem.memfree`` 改名为 ``mem.memusable``, 并分拆了 ``mem.free`` ``mem.cached`` ``mem.buffers``

修复 BUG：
1. 修复了一个插件调度的极端情况导致 agent crash 的 bug
2. 修复了一个 agent 无法更新规则仓库的 bug
3. 修复了部署描述符错误导致丢失数据的 bug
4. 修复了几处安装脚本的 bug
