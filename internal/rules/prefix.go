package rules

import (
	"cmp"
	"fmt"
	"net/netip"
	"slices"
)

type PrefixSet []netip.Prefix

func (set PrefixSet) Contains(address netip.Addr) bool {
	for _, prefix := range set {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func (set PrefixSet) Texts() []string {
	result := make([]string, len(set))
	for index, prefix := range set {
		result[index] = prefix.String()
	}
	return result
}

func (set PrefixSet) Optimized() PrefixSet {
	if len(set) == 0 {
		return nil
	}
	sorted := make(PrefixSet, len(set))
	for index, prefix := range set {
		if !prefix.IsValid() {
			panic("invalid IP prefix")
		}
		sorted[index] = prefix.Masked()
	}
	slices.SortFunc(sorted, comparePrefix)

	stack := sorted[:0]
	for _, prefix := range sorted {
		if len(stack) > 0 && stack[len(stack)-1].Contains(prefix.Addr()) {
			continue
		}
		stack = append(stack, prefix)
		for len(stack) >= 2 {
			left, right := stack[len(stack)-2], stack[len(stack)-1]
			if left.Addr().BitLen() != right.Addr().BitLen() || left.Bits() != right.Bits() || left.Bits() == 0 {
				break
			}
			parent := netip.PrefixFrom(left.Addr(), left.Bits()-1).Masked()
			if parent != netip.PrefixFrom(right.Addr(), right.Bits()-1).Masked() {
				break
			}
			stack = stack[:len(stack)-1]
			stack[len(stack)-1] = parent
		}
	}
	return stack
}

func (set PrefixSet) Excluding(earlier PrefixSet) PrefixSet {
	base := set.Optimized()
	cuts := earlier.Optimized()
	if len(base) == 0 || len(cuts) == 0 {
		return base
	}
	result := make(PrefixSet, 0, len(base))
	cutIndex := 0
	for _, prefix := range base {
		subtractSorted(prefix, cuts, &cutIndex, &result)
	}
	return result
}

func subtractSorted(current netip.Prefix, cuts PrefixSet, cutIndex *int, out *PrefixSet) {
	for *cutIndex < len(cuts) && !current.Overlaps(cuts[*cutIndex]) && cuts[*cutIndex].Addr().Less(current.Addr()) {
		*cutIndex++
	}
	if *cutIndex >= len(cuts) || !current.Overlaps(cuts[*cutIndex]) {
		*out = append(*out, current)
		return
	}
	if cuts[*cutIndex].Bits() <= current.Bits() {
		return
	}
	left, right := prefixChildren(current)
	subtractSorted(left, cuts, cutIndex, out)
	subtractSorted(right, cuts, cutIndex, out)
}

func prefixChildren(prefix netip.Prefix) (netip.Prefix, netip.Prefix) {
	bits := prefix.Bits()
	if bits < 0 || bits >= prefix.Addr().BitLen() {
		panic(fmt.Sprintf("prefix %s cannot be split", prefix))
	}
	childBits := bits + 1
	left := netip.PrefixFrom(prefix.Addr(), childBits).Masked()
	right := netip.PrefixFrom(setAddressBit(prefix.Addr(), bits), childBits).Masked()
	return left, right
}

func setAddressBit(address netip.Addr, bit int) netip.Addr {
	if address.Is4() {
		bytes := address.As4()
		bytes[bit/8] |= 1 << (7 - bit%8)
		return netip.AddrFrom4(bytes)
	}
	bytes := address.As16()
	bytes[bit/8] |= 1 << (7 - bit%8)
	return netip.AddrFrom16(bytes)
}

func comparePrefix(left, right netip.Prefix) int {
	return cmp.Or(left.Addr().Compare(right.Addr()), cmp.Compare(left.Bits(), right.Bits()))
}
