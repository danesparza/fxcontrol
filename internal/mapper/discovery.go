// Package mapper converts network advertisements into API models.
package mapper

import (
	"slices"
	"strings"

	"github.com/danesparza/fxcontrol/internal/model"
	"github.com/grandcat/zeroconf"
)

// DiscoveryService accepts the versioned FX advertisement shared by FX services.
func DiscoveryService(entry *zeroconf.ServiceEntry) (model.Service, bool) {
	if entry == nil || entry.TTL == 0 || entry.Port < 1 || entry.Port > 65535 {
		return model.Service{}, false
	}
	txt := make(map[string]string)
	for _, record := range entry.Text {
		key, value, ok := strings.Cut(record, "=")
		if ok {
			txt[key] = value
		}
	}
	if txt["txtvers"] != "1" || txt["id"] == "" || txt["api"] != "v1" || txt["scheme"] != "http" || txt["path"] != "/v1" {
		return model.Service{}, false
	}
	switch txt["service"] {
	case "fxaudio", "fxpixel", "fxdmx", "fxtrigger":
	default:
		return model.Service{}, false
	}
	addresses := make([]string, 0, len(entry.AddrIPv4)+len(entry.AddrIPv6))
	for _, ip := range entry.AddrIPv4 {
		if ip != nil {
			addresses = append(addresses, ip.String())
		}
	}
	for _, ip := range entry.AddrIPv6 {
		if ip != nil {
			addresses = append(addresses, ip.String())
		}
	}
	slices.Sort(addresses)
	addresses = slices.Compact(addresses)
	return model.Service{ID: txt["id"], Name: entry.Instance, Service: txt["service"], Host: entry.HostName, Port: entry.Port, Addresses: addresses, API: txt["api"], Scheme: txt["scheme"], Path: txt["path"]}, true
}
