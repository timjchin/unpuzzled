package cli

// https://github.com/golang/go/wiki/SliceTricks#reversing
func reverseStringSlice(a []string) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}
