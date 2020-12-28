package aggregate

import (
	"github.com/yl2chen/cidranger"
	"net"
	"sort"
)

func removeContained(pfxs []*net.IPNet) ([]*net.IPNet, error) {
	// Sort the supplied prefixes by the length of their prefixes.
	sort.Slice(pfxs, func(i, j int) bool {
		iLen, iFamily := pfxs[i].Mask.Size()
		jLen, jFamily := pfxs[j].Mask.Size()
		return iLen < jLen || iFamily < jFamily
	})

	// Sequentially test for the presence each (sorted) prefix in a ranger (tree), and if it is not already covered,
	// then add it into the tree so that longer prefixes are not needlessly added.
	ranger := cidranger.NewPCTrieRanger()
	for _, pfx := range pfxs {
		exists, err := ranger.Contains(pfx.IP)
		if err != nil {
			return nil, err
		}

		// Does the network address already exist in the ranger? If so, no need to add it.
		if exists {
			continue
		}

		// As the network address does not exist, add the prefix to the ranger.
		if err := ranger.Insert(cidranger.NewBasicRangerEntry(*pfx)); err != nil {
			return nil, err
		}
	}

	// Extract the networks out of the completed ranger.
	ipv4, err := ranger.CoveredNetworks(*cidranger.AllIPv4)
	if err != nil {
		return nil, err
	}
	ipv6, err := ranger.CoveredNetworks(*cidranger.AllIPv6)
	if err != nil {
		return nil, err
	}

	// Form the results into something useful to our caller.
	result := make([]*net.IPNet, 0, len(ipv4)+len(ipv6))
	for _, ipNet := range append(ipv4, ipv6...) {
		cidr := ipNet.Network()
		result = append(result, &cidr)
	}

	return result, nil
}

func mergeAdjacent(pfxs []*net.IPNet) []*net.IPNet {
	// Track modifications, keep running until a run completes with no modifications taking place.
	mod := true
	for mod == true {
		mod = false
		for i := 0; i < len(pfxs); i++ {
			// Skip the last element, as nothing to compare it with.
			if i == len(pfxs)-1 {
				break
			}

			// Are the prefix lengths (and address families) identical? If not, bail early.
			iLen, iFamily := pfxs[i].Mask.Size()
			jLen, jFamily := pfxs[i+1].Mask.Size()
			if iLen != jLen || iFamily != jFamily {
				continue
			}

			// Make the "left" prefix one size shorter, and see if the "right" prefix now fits within it. If it does,
			// remove the "right" prefix, and replace the "left" prefix with the newly enlarged prefix.
			pfx := &net.IPNet{
				IP:   pfxs[i].IP,
				Mask: net.CIDRMask(iLen-1, iFamily),
			}
			if pfx.Contains(pfxs[i+1].IP) {
				pfxs[i] = pfx

				// Handle the case where we are near the end of the slice, without panicking.
				if len(pfxs)-i > 1 {
					pfxs = append(pfxs[:i+1], pfxs[i+2:]...)
				} else {
					pfxs = pfxs[:i+1]
				}

				// Mark that modifications have been made.
				mod = true
			}
		}
	}
	return pfxs
}

// IPNets takes a slice of CIDR prefixes and aggregates the prefixes to the smallest possible set of prefixes that
// covers the exact same set of addresses.
func IPNets(pfxs []*net.IPNet) ([]*net.IPNet, error) {
	contained, err := removeContained(pfxs)
	if err != nil {
		return nil, err
	}
	return mergeAdjacent(contained), nil
}

// Strings is a convenience function that accepts a slice of CIDR prefix strings instead of net.IPNet structs.
func Strings(pfxs []string) ([]string, error) {
	ipNets := make([]*net.IPNet, 0, len(pfxs))
	for _, pfx := range pfxs {
		_, ipNet, err := net.ParseCIDR(pfx)
		if err != nil {
			return nil, err
		}
		ipNets = append(ipNets, ipNet)
	}

	ipNets, err := IPNets(ipNets)
	if err != nil {
		return nil, err
	}

	ipNetStrs := make([]string, 0, len(ipNets))
	for _, ipNet := range ipNets {
		ipNetStrs = append(ipNetStrs, ipNet.String())
	}

	return ipNetStrs, nil
}