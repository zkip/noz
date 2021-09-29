package utils

import (
	"net"

	mapset "github.com/deckarep/golang-set"
)

func ResolveIP(hostname string) []string {
	addrs, _ := net.LookupHost(hostname)
	for i := 0; i < len(addrs); i++ {
		if hostname == "localhost" {
			addrs[i] = "127.0.0.1"
		}
	}
	return addrs
}

func IsAffinityHost(destHostname string, originHostname string) bool {

	destAddrs := ResolveIP(destHostname)
	originAddrs := ResolveIP(originHostname)
	destAddrsSet := mapset.NewSet(ToInterfaceSlice(destAddrs)...)
	originAddrsSet := mapset.NewSet(ToInterfaceSlice(originAddrs)...)
	result := destAddrsSet.Intersect(originAddrsSet)

	return result.Cardinality() > 0
}
