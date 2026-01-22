package dyalpm

// VerCmp compares two version strings according to libalpm version comparison rules.
// Returns <0 if v1 < v2, 0 if v1 == v2, >0 if v1 > v2.
func VerCmp(v1, v2 string) int {
	if v1 == v2 {
		return 0
	}

	// Split into epoch:version-rel
	ae, av, ar := splitVersion(v1)
	be, bv, br := splitVersion(v2)

	// Compare epochs
	if ae != be {
		return compareNumericString(ae, be)
	}

	// Compare versions
	ret := rpmvercmpGo(av, bv)
	if ret != 0 {
		return ret
	}

	// Compare releases (only if both have releases)
	if ar != "" && br != "" {
		return rpmvercmpGo(ar, br)
	}

	return 0
}

// splitVersion splits a version string into epoch, version, and release components.
// This follows the exact logic from libalpm's parseEVR function.
func splitVersion(v string) (epoch, version, rel string) {
	epoch = "0"
	version = v
	rel = ""

	// Find epoch: skip leading digits, then check if next char is ':'
	// This matches C: while (*s && isdigit(*s)) s++;
	i := 0
	for i < len(v) && isDigit(v[i]) {
		i++
	}

	// If we found ':' right after the digits, that's the epoch separator
	if i < len(v) && v[i] == ':' {
		if i > 0 {
			epoch = v[:i]
		} else {
			// Empty epoch before ':' means "0"
			epoch = "0"
		}
		version = v[i+1:]
	}
	// Otherwise epoch stays "0" and version is the whole string

	// Find release (after last '-')
	for i := len(version) - 1; i >= 0; i-- {
		if version[i] == '-' {
			rel = version[i+1:]
			version = version[:i]
			break
		}
	}

	return epoch, version, rel
}

func rpmvercmpGo(a, b string) int {
	if a == b {
		return 0
	}

	one := 0 // index into a
	two := 0 // index into b

	// Loop while BOTH strings have content remaining
	for one < len(a) && two < len(b) {
		// Save positions before skipping separators
		ptr1 := one
		ptr2 := two

		// Skip non-alphanumeric characters
		for one < len(a) && !isAlnum(a[one]) {
			one++
		}
		for two < len(b) && !isAlnum(b[two]) {
			two++
		}

		// If we ran to the end of either, we are finished with the loop
		if one >= len(a) || two >= len(b) {
			break
		}

		// If the separator lengths were different, we are also finished
		if (one - ptr1) != (two - ptr2) {
			if (one - ptr1) < (two - ptr2) {
				return -1
			}
			return 1
		}

		// Save start positions of this segment
		ptr1 = one
		ptr2 = two

		var isNum bool
		// Grab first completely alpha or completely numeric segment
		if isDigit(a[ptr1]) {
			isNum = true
			for one < len(a) && isDigit(a[one]) {
				one++
			}
			for two < len(b) && isDigit(b[two]) {
				two++
			}
		} else {
			isNum = false
			for one < len(a) && isAlpha(a[one]) {
				one++
			}
			for two < len(b) && isAlpha(b[two]) {
				two++
			}
		}

		// Extract segments
		segA := a[ptr1:one]
		segB := b[ptr2:two]

		// This cannot happen, as we previously tested to make sure that
		// the first string has a non-null segment
		if len(segA) == 0 {
			return -1 // arbitrary
		}

		// Take care of the case where the two version segments are
		// different types: one numeric, the other alpha (i.e. empty)
		// numeric segments are always newer than alpha segments
		if len(segB) == 0 {
			if isNum {
				return 1
			}
			return -1
		}

		if isNum {
			// Throw away any leading zeros - it's a number, right?
			segA = stripLeadingZeros(segA)
			segB = stripLeadingZeros(segB)

			// Whichever number has more digits wins
			if len(segA) > len(segB) {
				return 1
			}
			if len(segB) > len(segA) {
				return -1
			}
		}

		if segA < segB {
			return -1
		}
		if segA > segB {
			return 1
		}
	}

	if one >= len(a) && two >= len(b) {
		return 0
	}

	oneEmpty := one >= len(a)
	twoEmpty := two >= len(b)

	oneIsAlpha := !oneEmpty && isAlpha(a[one])
	twoIsAlpha := !twoEmpty && isAlpha(b[two])

	if (oneEmpty && !twoIsAlpha) || oneIsAlpha {
		return -1
	}
	return 1
}

// stripLeadingZeros strips leading zeros but keeps at least one character
func stripLeadingZeros(s string) string {
	i := 0
	for i < len(s)-1 && s[i] == '0' {
		i++
	}
	return s[i:]
}

// compareNumericString compares two numeric strings
func compareNumericString(a, b string) int {
	// Strip leading zeros
	a = stripLeadingZeros(a)
	b = stripLeadingZeros(b)

	// Compare lengths first
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}

	// Same length, compare lexicographically
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func isAlpha(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isAlnum(c byte) bool { return isDigit(c) || isAlpha(c) }
