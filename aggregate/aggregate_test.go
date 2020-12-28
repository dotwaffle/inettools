package aggregate

import (
	"github.com/google/go-cmp/cmp"
	"net"
	"testing"
)

func TestAggregate(t *testing.T) {
	tests := map[string]struct{
		input []string
		want []string
	}{
		"Nil": {
			input: nil,
			want: []string{},
		},
		"Empty": {
			input: []string{},
			want: []string{},
		},
		"Contained": {
			input: []string{
				"0.0.0.0/0",
				"192.0.2.0/24",
			},
			want: []string{
				"0.0.0.0/0",
			},
		},
		"Duplicates": {
			input: []string{
				"192.0.2.0/24",
				"192.0.2.0/24",
			},
			want: []string{
				"192.0.2.0/24",
			},
		},
		"MergedOnce": {
			input: []string{
				"192.0.2.0/25",
				"192.0.2.128/25",
			},
			want: []string{
				"192.0.2.0/24",
			},
		},
		"MergedLots": {
			input: []string{
				"192.0.2.0/32",
				"192.0.2.1/32",
				"192.0.2.2/32",
				"192.0.2.3/32",
				"192.0.2.4/32",
				"192.0.2.5/32",
				"192.0.2.6/32",
				"192.0.2.7/32",
			},
			want: []string{
				"192.0.2.0/29",
			},
		},
		"MergedHole": {
			input: []string{
				"192.0.2.0/32",
				"192.0.2.1/32",
				"192.0.2.2/32",
				"192.0.2.3/32",
				// skip 192.0.2.4/32 to create a hole that should be preserved.
				"192.0.2.5/32",
				"192.0.2.6/32",
				"192.0.2.7/32",
			},
			want: []string{
				"192.0.2.0/30",
				"192.0.2.5/32",
				"192.0.2.6/31",
			},
		},
		"HostAddresses": {
			// A collection of host addresses (with masks) rather than network addresses.
			input: []string{
				"192.0.2.1/29",
				"192.0.2.2/29",
				"192.0.2.9/29",
			},
			want: []string{
				"192.0.2.0/28",
			},
		},
		"IPv6": {
			input: []string{
				"::/0",
				"2001:db8::/32",
			},
			want: []string{
				"::/0",
			},
		},
		"IPv4+IPv6": {
			// IPv4 always gets printed first, because of the sorting done prefers number of address bits before length.
			input: []string{
				"192.0.2.0/25",
				"192.0.2.128/25",
				"2001:db8::/32",
				"2001:db8::/48",
			},
			want: []string{
				"192.0.2.0/24",
				"2001:db8::/32",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name+"/IPNets", func(t *testing.T) {
			ipNets := make([]*net.IPNet, 0, len(tc.input))
			// Convert from string to net.IPNet for function.
			for _, ipNetStr := range tc.input {
				_, ipNet, err := net.ParseCIDR(ipNetStr)
				if err != nil {
					t.Fatalf("input: %s produced err: %v", ipNetStr, err)
				}
				ipNets = append(ipNets, ipNet)
			}

			got, err := IPNets(ipNets)
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			// Convert output back to strings for comparison.
			gotStrs := make([]string, 0, len(got))
			for _, gotIPNet := range got {
				gotStrs = append(gotStrs, gotIPNet.String())
			}

			diff := cmp.Diff(tc.want, gotStrs)
			if diff != "" {
				t.Fatalf("%v", diff)
			}
		})
		t.Run(name+"/Strings", func(t *testing.T) {
			got, err := Strings(tc.input)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf("%v", diff)
			}
		})
	}
}

func benchmarkIPNets(l int, b *testing.B) {
	pfxs := make([]*net.IPNet, 1<<(32-l))
	switch {
	case l >= 24:
		for i := 0; i <= 1<<(32-l)-1; i++ {
			pfxs[i] = &net.IPNet{
				IP:   net.IPv4(10, 0, 0, byte(i)),
				Mask: net.CIDRMask(32, 32),
			}
		}
	case l >= 16 && l < 24:
		for i := 0; i <= 255; i++ {
			for j := 0; j <= 1<<(24-l)-1; j++ {
				pfxs[(j*256)+i] = &net.IPNet{
					IP:   net.IPv4(10, 0, byte(j), byte(i)),
					Mask: net.CIDRMask(32, 32),
				}
			}
		}
	default:
		b.Fatalf("length too long to produce reasonable results: %d", l)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		if _, err := IPNets(pfxs); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIPNets16(b *testing.B) { benchmarkIPNets(16, b) }
func BenchmarkIPNets17(b *testing.B) { benchmarkIPNets(17, b) }
func BenchmarkIPNets18(b *testing.B) { benchmarkIPNets(18, b) }
func BenchmarkIPNets19(b *testing.B) { benchmarkIPNets(19, b) }
func BenchmarkIPNets20(b *testing.B) { benchmarkIPNets(20, b) }
func BenchmarkIPNets21(b *testing.B) { benchmarkIPNets(21, b) }
func BenchmarkIPNets22(b *testing.B) { benchmarkIPNets(22, b) }
func BenchmarkIPNets23(b *testing.B) { benchmarkIPNets(23, b) }
func BenchmarkIPNets24(b *testing.B) { benchmarkIPNets(24, b) }
func BenchmarkIPNets25(b *testing.B) { benchmarkIPNets(25, b) }
func BenchmarkIPNets26(b *testing.B) { benchmarkIPNets(26, b) }
func BenchmarkIPNets27(b *testing.B) { benchmarkIPNets(27, b) }
func BenchmarkIPNets28(b *testing.B) { benchmarkIPNets(28, b) }
func BenchmarkIPNets29(b *testing.B) { benchmarkIPNets(29, b) }
func BenchmarkIPNets30(b *testing.B) { benchmarkIPNets(30, b) }
func BenchmarkIPNets31(b *testing.B) { benchmarkIPNets(31, b) }
func BenchmarkIPNets32(b *testing.B) { benchmarkIPNets(32, b) }
