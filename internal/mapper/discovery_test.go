package mapper

import (
	"net"
	"reflect"
	"testing"

	"github.com/grandcat/zeroconf"
)

func TestDiscoveryService(t *testing.T) {
	for _, kind := range []string{"fxaudio", "fxpixel", "fxdmx", "fxtrigger"} {
		t.Run(kind, func(t *testing.T) {
			entry := zeroconf.NewServiceEntry("Porch", "_fx._tcp", "local.")
			entry.HostName = "stage.local."
			entry.Port = 3030
			entry.TTL = 120
			entry.Text = []string{"txtvers=1", "service=" + kind, "id=installation=1", "api=v1", "scheme=http", "path=/v1", "unknown=value"}
			entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.10"), net.ParseIP("192.168.1.10")}
			entry.AddrIPv6 = []net.IP{net.ParseIP("2001:db8::1")}
			got, ok := DiscoveryService(entry)
			if !ok || got.Service != kind || got.ID != "installation=1" || got.Name != "Porch" || got.Host != entry.HostName || got.Port != 3030 || got.API != "v1" || got.Scheme != "http" || got.Path != "/v1" || !reflect.DeepEqual(got.Addresses, []string{"192.168.1.10", "2001:db8::1"}) {
				t.Fatalf("unexpected mapping: %+v, %v", got, ok)
			}
		})
	}
	for _, tc := range []struct {
		name   string
		mutate func(*zeroconf.ServiceEntry)
	}{
		{"goodbye", func(e *zeroconf.ServiceEntry) { e.TTL = 0 }},
		{"port zero", func(e *zeroconf.ServiceEntry) { e.Port = 0 }},
		{"port overflow", func(e *zeroconf.ServiceEntry) { e.Port = 65536 }},
		{"unknown service", func(e *zeroconf.ServiceEntry) { e.Text[1] = "service=printer" }},
		{"controller", func(e *zeroconf.ServiceEntry) { e.Text[1] = "service=fxcontrol" }},
		{"missing id", func(e *zeroconf.ServiceEntry) { e.Text[2] = "invalid" }},
		{"unknown version", func(e *zeroconf.ServiceEntry) { e.Text[0] = "txtvers=2" }},
		{"unsupported API", func(e *zeroconf.ServiceEntry) { e.Text[3] = "api=v2" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entry := zeroconf.NewServiceEntry("test", "_fx._tcp", "local.")
			entry.Port, entry.TTL = 3030, 120
			entry.Text = []string{"txtvers=1", "service=fxaudio", "id=one", "api=v1", "scheme=http", "path=/v1"}
			tc.mutate(entry)
			if got, ok := DiscoveryService(entry); ok {
				t.Fatalf("accepted invalid entry: %+v", got)
			}
		})
	}
	if _, ok := DiscoveryService(nil); ok {
		t.Fatal("accepted nil entry")
	}
}
