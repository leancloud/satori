package g

// changelog:
// 3.1.3: code refactor
// 3.1.4: bugfix ignore configuration
// 5.0.0: 支持通过配置控制是否开启/run接口；收集udp流量数据；du某个目录的大小
// 5.1.0: 同步插件的时候不再使用checksum机制
// 5.1.1: 修复往多个transfer发送数据的时候crash的问题
// 6.0.0: 配合 satori master
// 7.0.0: tags 改成 map[string]string 了，不兼容
// 7.0.3: 优化与 transfer 之间的数据传输
// 7.0.4: 报告执行失败的插件
// 7.0.5: 修复与 master 断线之后不能恢复的bug
// 7.0.6: 增加 noBuiltIn 参数，打开后只采集插件参数。
// 7.0.7: 增加了 /v1/ping ，用来检测 agent 存活
// 7.0.8: 支持 long running 的插件
// 7.0.9: 补全插件上报的事件中的 endpoint
// 7.0.10: 新增插件签名功能，无签名的规则仓库不会主动切换过去。
// 7.0.11: 新增 agent 自更新功能
// 7.0.12: 增加 alternative key file 功能，管理签名 key 更方便了
// 7.0.13: 自更新的 binary 放到仓库外，配置文件格式修改（YAML)
// 7.0.14: 修复一个严重的bug
// 7.0.15: 增加 cgroups 限制
// 7.0.16: 小 bug 修复

const (
	VERSION = "7.0.16"
)
