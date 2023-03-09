package ip_manager

import (
	"testing"

	"github.com/lbryio/lbry.go/v2/extras/stop"
)

func TestAll(t *testing.T) {
	stopGroup := stop.New()
	pool, err := GetIPPool(stopGroup)
	if err != nil {
		t.Fatal(err)
	}
	ip, err := pool.GetIP("test")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
	pool.ReleaseIP(ip)
	ip2, err := pool.GetIP("test")
	if err != nil {
		t.Fatal(err)
	}
	if ip == ip2 && len(pool.ips) > 1 {
		t.Fatalf("the same IP was returned twice! %s, %s", ip, ip2)
	}
	t.Log(ip2)
	pool.ReleaseIP(ip2)

	for range pool.ips {
		_, err = pool.GetIP("test")
		if err != nil {
			t.Fatal(err)
		}
	}
	next, err := pool.nextIP("test")
	if err != nil {
		t.Logf("%s", err.Error())
	} else {
		t.Fatal(next)
	}
}
