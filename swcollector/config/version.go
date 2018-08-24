package config

// changelog:
// 3.1.3: code refactor
// 3.1.4: bugfix ignore configuration
// 3.1.5: more sw support, DisplayByBit cfg
// 3.1.6
// 3.2.0: more sw support, fix ping bug, add ifOperStatus,ifBroadcastPkt,ifMulticastPkt
// 3.2.1 add Discards,Error,UnknownProtos,QLen，fix some bugs
// 3.2.1.1 debugmetric support multi endpoints and metrics
// 3.2.1.2 gosnmp use getnext to walk snmp
// 4.0.0 caculate counter type on swcollect local,add speedpercent
// 4.0.1 fix sometimes ifstat pannic
// 4.0.2 fix speedpercent bug
// 4.0.4 add lock on map;add limconn for switch snmp request
// 4.0.5 add custom metric,custom host
// 4.0.6.1 fix channal closed bug
// 4.0.6.2 fix Vendor bug;add remote config api
// 4.0.6.3 fix bugs
// --------------
// 5.0.0 first ported version, complete rewrite
const (
	VERSION = "5.0.0"
)
