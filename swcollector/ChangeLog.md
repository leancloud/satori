# Changelog #
## 5.0.0 ##
首个移植版本
1. 移除了 HTTP 功能
2. 移除了 hosts 功能，添加反向解析功能

## ----- 移植分割线 -----

## 4.0.6.3 ##
#### bug修复 ####
1. 现在当 tag 为空时，debug 应该会正确打印对应的日志了
2. 修复了一个当自定义 oid 超过 2 项，且 oid 不正确无法采集到信息时，swcollector 会异常崩溃的 bug
3. 修复了一个当 speedlimit 使用指定值（而非自动采用接口速率 ifSpeed)作为限制时，接口的 speedPercent 无法正确采集的 bug
#### 改进 ####
1. 配置热重载模式调整，现在 reload 后会在下一个采集周期时重载配置。并清空 AliveIp，因此配置中移除的 IP 可以正确生效了
2. swcollector 自带的 http 页，现在应该能更快速的显示了
3. swcollector 自带的 http 页，现在支持更多交换机的 sysmodle 了
4. cpu 和 mem 现在对于老版本的 H3C (H3C Comware Platform Software Comware software, Version 3.10)，应该也能正确采集到了


# Changelog #
## 4.0.6.2 ##
#### 新功能 ####
1. 增加了动态重载配置的功能，详见 [README](https://github.com/gaochao1/swcollector/blob/master/README.md})
#### bug修复 ####
1. 现在当 tag 为空时，debug 的时候应该不会重复的打印日志了
#### 改进 ####
1. 现在自定义 Oid 采集时，可以支持 string 类型的返回了。系统会强制转换成 float64 上报，如果转换出错则抛出错误
2. 现在对交换机类型的判断时，也会采取重试（由配置中的 retry 决定重试次数）来规避偶发性的异常了。
3. 现在对 Linux 的 snmp 监控能够采集到 cpu 和 mem 了


## 4.0.6.1 ##
#### bug修复 ####
1. 现在当采集异常 Channel 关闭时，应该会正常的抛弃而不会给 transfer 上报一个空的 endpoint 了

## 4.0.6 ##
#### 新功能 ####
1. 增加接口速率的采集
	* switch.if.Speed
2. 增加接口流量百分比的采集
	* switch.if.InSpeedPercent
	* switch.if.OutSpeedPercent
3. 增加 dell 交换机的内置 cpu/mem 采集
4. 现在支持自定义 oid 的采集了
5. 现在支持自定义交换机 host，以 host 作为 endpoint 上报
6. 现在支持地址段的方式配置采集列表，例如 ```"192.168.56.102-192.168.56.120"```
7. 现在对于单台交换机的多个指标采集，也支持并发限制了。以减少对一些 snmp 响应较低的交换机采集时，由于并发太多产生的超时问题。
 
#### 改进 ####
1. Counter 类型的数据上报逻辑大幅更改。现在 swcollector 将在本地计算出相应的数值，再以 Gauge 类型上报。如果出现异常的数据，则在本地直接抛弃。因此最终呈现的绘图至多只会出现断点，而不再会出现极端的异常图形。
2. 优化了 gosnmp 的端口采集，现在 gosnmp 端口采集异常的超时情况应该大幅度降低了
3. 现在如果并发采集的某个 goroutine 的耗时超过了采集周期，则该 goroutine 会超时退出，避免异常时大量 goroutine 耗尽资源
4. 移除了对 Cisco_ASA 的防火墙连接数内置采集，此类需求今后可通过自定义 oid 方式采集。

#### bug修复 ####
1. 现在当 cpu 和 mem 采集异常的时候，应该能正确的抛弃。而不是上报一个 0 了


## 3.2.1.1 ##
#### 新功能 ####
1. debugmetric 现在支持配置多个 endpoint 和 metric 了
## 3.2.1 ##
#### 新功能 ####
1. 增加接口丢包数量的采集
	* IfInDiscards
	* IfOutDiscards

2. 增加接口错包数量的采集
	* IfInErrors
	* IfOutErros
	
3. 增加接口由于未知或不支持协议造成的丢包数量采集
	* IfInUnknownProtos
	
4. 增加接口输出队列中的包数量采集
	* IfOutQLen

5. 现在能够通过 debugmetric 配置选项，配置想要 debug 的 metric 了。配置后日志中会具体打印该条 metric 的日志

6. 现在能够通过 gosnmp 的配置选项,选择采用 gosnmp 还是 snmpwalk 进行数据采集了。两者效率相仿，snmpwalk稍微多占用点系统资源

#### 改进 ####
1. 优化了 gosnmp 的端口采集，略微控制了一下并发速率，现在 gosnmp 采集端口超时的概率，应该有所降低了
2. 代码优化，删除了部分无关代码（比如 hbs 相关的部分……)
3. 部分日志的输出可读性更强了

#### bug修复 ####
1. 修复了一个广播报文采集不正确的 bug
2. 修复了一个老版本思科交换机，CPU 内存采集不正确的 bug
3. 修复了一些偶尔进程崩溃的 bug

## 3.2.0 ##
#### 新功能 ####
1. 增加接口广播包数量的采集
	* IfHCInBroadcastPkts
	* IfHCOutBroadcastPkts

2. 增加接口组播包数量的采集
	* IfHCInMulticastPkts
	* IfHCOutMulticastPkts

3. 增加接口状态的采集
	* IfOperStatus(1 up, 2 down, 3 testing, 4 unknown, 5 dormant, 6 notPresent, 7 lowerLayerDown)

4. 内置了更多交换机型号的 CPU， 内存的 OID 和计算方式。（锐捷，Juniper, 华为， 华三的一些型号等)

PS: 虽然 if 采集是并发的，不过采集项开的太多还是可能会影响 snmp 的采集效率，尤其是华为等 snmp 返回比较慢的交换机…………故谨慎选择，按需开启。

#### 改进 ####
1. 解决了 if 采集乱序的问题，现在即便使用 gosnmp 采集返回乱序也可以正确处理了。已测试过的华为型号现在均使用 gosnmp 采集。（v5.13，v5.70，v3.10）
2. 现在 log 中 打印 panic 信息的时候，应该会带上具体的 ip 地址了。
3. 现在默认采集 bit 单位的网卡流量了。
4. 去掉了默认配置文件里的 hostname 和 ip 选项，以免产生歧义，反正也没什么用…………
5. 修改默认 http 端口为 1989，避免和 agent 的端口冲突。

PS: func/swifstat.go 151行的注释代码，会在 debug 模式下打印具体的 ifstat 输出。如果交换机采集数据出现不准确的情况，可开启这段代码来进行排查。

#### bug修复 ####
1. 修复了在并发 ping 的情况下，即便 ip 地址不通，也有小概率 ping 通地址的 bug。（很神奇是不是……反正在我这里有出现这现象。。。）。方案是替换为 [go-fastping](https://github.com/tatsushid/go-fastping) 来做 ping 探测，通过 fastPingMode 配置选项开启。
2. 修复了思科 ASA-5585 9.1 和 9.2 两个版本 cpu, memory 的 oid 不一致带来的采集问题。（这坑爹玩意!)。现在应该可以根据他的版本号来选择不同的 oid 进行采集了。
