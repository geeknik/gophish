package dialer

import (
	"fmt"
	"net"
	"strings"
	"syscall"
	"testing"
)

func TestDefaultBlocksAllInternalNetworks(t *testing.T) {
	control := restrictedControl([]*net.IPNet{})
	conn := new(syscall.RawConn)

	internalHosts := []string{
		"127.0.0.1",
		"10.0.0.1",
		"172.16.0.1",
		"192.168.1.1",
		"169.254.169.254",
	}

	for _, host := range internalHosts {
		got := control("tcp4", fmt.Sprintf("%s:80", host), *conn)
		if got == nil || !strings.Contains(got.Error(), "upstream connection denied") {
			t.Fatalf("default should block %s, got: %v", host, got)
		}
	}
}

func TestDefaultAllowsExternalHosts(t *testing.T) {
	control := restrictedControl([]*net.IPNet{})
	conn := new(syscall.RawConn)

	externalHosts := []string{
		"1.1.1.1",
		"8.8.8.8",
		"93.184.216.34",
	}

	for _, host := range externalHosts {
		got := control("tcp4", fmt.Sprintf("%s:80", host), *conn)
		if got != nil {
			t.Fatalf("default should allow external host %s, got: %v", host, got)
		}
	}
}

func TestCustomAllow(t *testing.T) {
	host := "127.0.0.1"
	_, ipRange, _ := net.ParseCIDR(fmt.Sprintf("%s/32", host))
	allowed := []*net.IPNet{ipRange}
	control := restrictedControl(allowed)
	conn := new(syscall.RawConn)
	got := control("tcp4", fmt.Sprintf("%s:80", host), *conn)
	if got != nil {
		t.Fatalf("error dialing allowed host. got %v", got)
	}
}

func TestCustomDeny(t *testing.T) {
	host := "127.0.0.1"
	_, ipRange, _ := net.ParseCIDR(fmt.Sprintf("%s/32", host))
	allowed := []*net.IPNet{ipRange}
	control := restrictedControl(allowed)
	conn := new(syscall.RawConn)
	expected := fmt.Errorf("upstream connection denied to internal host at %s", host)
	got := control("tcp4", "192.168.1.2:80", *conn)
	if !strings.Contains(got.Error(), "upstream connection denied") {
		t.Fatalf("unexpected error dialing denylisted host. expected %v got %v", expected, got)
	}
}

func TestSingleIP(t *testing.T) {
	orig := DefaultDialer.AllowedHosts()
	host := "127.0.0.1"
	DefaultDialer.SetAllowedHosts([]string{host})
	control := DefaultDialer.Dialer().Control
	conn := new(syscall.RawConn)
	expected := fmt.Errorf("upstream connection denied to internal host at %s", host)
	got := control("tcp4", "192.168.1.2:80", *conn)
	if !strings.Contains(got.Error(), "upstream connection denied") {
		t.Fatalf("unexpected error dialing denylisted host. expected %v got %v", expected, got)
	}

	host = "::1"
	DefaultDialer.SetAllowedHosts([]string{host})
	control = DefaultDialer.Dialer().Control
	conn = new(syscall.RawConn)
	expected = fmt.Errorf("upstream connection denied to internal host at %s", host)
	got = control("tcp4", "192.168.1.2:80", *conn)
	if !strings.Contains(got.Error(), "upstream connection denied") {
		t.Fatalf("unexpected error dialing denylisted host. expected %v got %v", expected, got)
	}

	// Test an allowed connection
	got = control("tcp4", fmt.Sprintf("[%s]:80", host), *conn)
	if got != nil {
		t.Fatalf("error dialing allowed host. got %v", got)
	}
	DefaultDialer.SetAllowedHosts(orig)
}

func TestAllowedInternalHostsConfig(t *testing.T) {
	orig := DefaultDialer.AllowedHosts()
	defer DefaultDialer.SetAllowedHosts(orig)

	DefaultDialer.SetAllowedHosts([]string{"192.168.1.0/24"})
	control := DefaultDialer.Dialer().Control
	conn := new(syscall.RawConn)

	got := control("tcp4", "192.168.1.50:80", *conn)
	if got != nil {
		t.Fatalf("allowed internal host should be permitted: %v", got)
	}

	got = control("tcp4", "192.168.2.50:80", *conn)
	if got == nil || !strings.Contains(got.Error(), "upstream connection denied") {
		t.Fatalf("non-allowed internal host should be blocked: %v", got)
	}
}
