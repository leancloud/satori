package sw

import (
	"log"

	"strings"
	"time"

	"github.com/gaochao1/gosnmp"
)

func SysModel(ip, community string, retry int, timeout int) (string, error) {
	vendor, err := SysVendor(ip, community, retry, timeout)
	method := "get"
	var oid string

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in sw.modelstat.go SysModel", r)
		}
	}()

	switch vendor {
	case "Cisco_NX", "Cisco", "Cisco_old", "Cisco_IOS_XR", "Cisco_IOS_XE", "Ruijie":
		method = "getnext"
		oid = "1.3.6.1.2.1.47.1.1.1.1.13"
	case "Huawei_ME60", "Huawei_V5.70", "Huawei_V5.130", "Huawei_V3.10":
		method = "getnext"
		oid = "1.3.6.1.2.1.47.1.1.1.1.2"
	case "H3c_V3.10", "H3C_S9500", "H3C", "H3C_V5", "H3C_V7", "Cisco_ASA":
		oid = "1.3.6.1.2.1.47.1.1.1.1.13"
		return getSwmodle(ip, community, oid, timeout, retry)
	case "Linux":
		return "Linux", nil
	default:
		return "", err
	}

	snmpPDUs, err := RunSnmp(ip, community, oid, method, timeout)

	if err == nil {
		for _, pdu := range snmpPDUs {
			return pdu.Value.(string), err
		}
	}

	return "", err

}

func getSwmodle(ip, community, oid string, timeout, retry int) (value string, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(ip+" Recovered in CPUtilization", r)
		}
	}()
	method := "getnext"
	oidnext := oid
	var snmpPDUs []gosnmp.SnmpPDU

	for {
		for i := 0; i < retry; i++ {
			snmpPDUs, err = RunSnmp(ip, community, oidnext, method, timeout)
			if len(snmpPDUs) > 0 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		oidnext = snmpPDUs[0].Name
		if strings.Contains(oidnext, oid) {
			if snmpPDUs[0].Value.(string) != "" {
				value = snmpPDUs[0].Value.(string)
				break
			}
		} else {
			break
		}

	}

	return value, err
}
