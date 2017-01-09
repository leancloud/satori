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
const (
	VERSION = "7.0.4"
)
