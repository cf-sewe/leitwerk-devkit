package gate

// Tier derivation for `verify --auto`: given a set of changed paths, pick the
// highest blast-radius tier so the gate runs exactly what the change warrants.
// Kept pure (no I/O) so it is unit-testable; the caller supplies the changed set
// (from a git diff) and runs the returned tier.

// RankTier maps a tier name to its position on the fixed blast-radius ladder
// T0 < T1 < T2. An unrecognized tier ranks ABOVE T2 so `--auto` errs toward more
// verification on a non-standard tiers file — it must never under-select.
func RankTier(tier string) int {
	switch tier {
	case "T0":
		return 0
	case "T1":
		return 1
	case "T2":
		return 2
	default:
		return 99
	}
}

// HighestTier returns the highest-blast-radius tier among the changed paths and
// the path that set it (empty when the result stays T0). Each path is classified
// with the same rule as the `tier` command — TierForPath, falling back to T1 on
// no match — and the maximum on the T0<T1<T2 ladder wins. An empty set yields T0
// (nothing changed is the lowest blast radius, and matches CI's initial value).
func HighestTier(t *Tiers, paths []string) (tier, deciding string) {
	tier = "T0"
	for _, p := range paths {
		pt, ok := t.TierForPath(p)
		if !ok {
			pt = "T1"
		}
		if RankTier(pt) > RankTier(tier) {
			tier, deciding = pt, p
		}
	}
	return tier, deciding
}
