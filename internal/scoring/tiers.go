package scoring

import "math"

// tierThresholdPct maps each non-Insufficient tier to its percentage
// threshold of max points (§3.1). Indexed by Tier; TierInsufficient (0)
// has no threshold (it's the fallback).
var tierThresholdPct = [...]float64{
	TierInsufficient: 0.0,
	TierSufficient:   0.70,
	TierGood:         0.76,
	TierVeryGood:     0.86,
	TierExcellent:    0.96,
}

// tierCutoff returns the inclusive lower-bound point value for a tier on
// a given max, applying §3.1's ceil-of-percentage rule. TierInsufficient
// always cuts off at 0.
func tierCutoff(t Tier, max Points) Points {
	if t == TierInsufficient {
		return 0
	}
	return Points(math.Ceil(tierThresholdPct[t] * float64(max)))
}

// BandFor returns the tier a point value falls into, given the maximum
// possible points for the unit being evaluated. Implements §3.1.
//
// Bands are percentage-based on the unit's own max, so a 25-point
// exercise and a 10-point exercise have different integer cutoffs.
//
// The rulebook publishes explicit cutoffs for the 25-point scale
// (24–25 Excellent, 22–23 Very Good, 19–21 Good, 18 Sufficient,
// 0–17 Insufficient). These are derived by:
//   - Excellent:    points >=  96% of max
//   - Very Good:    points >=  86% of max
//   - Good:         points >=  76% of max
//   - Sufficient:   points >=  70% of max
//   - Insufficient: otherwise
//
// Cutoffs are computed via math.Ceil of the percentage threshold,
// which reproduces the published 25-point table exactly.
//
// At small max values some tiers are mathematically unreachable.
// On a 5-point scale only Excellent (5), Good (4), and Insufficient
// (0–3) have integer outcomes — Very Good and Sufficient collapse.
// This is by design per §3.1 ("muzzle stability under social pressure
// is largely a binary competence at L1"). Callers should not "fix"
// BandFor to interpolate; the rulebook expects collapse.
//
// BandFor takes already-rounded Points, not float64. Rounding per
// §3.2 happens upstream via RoundPoints. This means a score that
// rounds from 17.5 to 18 is band-evaluated as 18, never as 17.5.
func BandFor(points, max Points) Tier {
	if max <= 0 {
		return TierInsufficient
	}
	for t := TierExcellent; t > TierInsufficient; t-- {
		if points >= tierCutoff(t, max) {
			return t
		}
	}
	return TierInsufficient
}

// BandCutoffs returns the inclusive lower-bound point value for each
// tier on a given max. Useful for rendering tier-reference tables on
// the tablet and for tests verifying the published rulebook tables.
//
// Tiers unreachable on the given max (e.g., TierVeryGood on a 5-point
// scale) are omitted from the returned map. Callers iterating the
// map should treat absence as "this tier cannot be earned at this max."
//
// Cutoffs are properties of the max value alone, not of any particular
// exercise. Two exercises with max=10 share cutoffs.
func BandCutoffs(max Points) map[Tier]Points {
	out := map[Tier]Points{TierInsufficient: 0}
	if max <= 0 {
		return out
	}
	// Walk from highest tier down. A tier is reachable only if its
	// cutoff is strictly less than the next-higher reachable tier's
	// cutoff — otherwise no integer point value lands in its band.
	prev := tierCutoff(TierExcellent, max)
	out[TierExcellent] = prev
	for t := TierVeryGood; t >= TierSufficient; t-- {
		c := tierCutoff(t, max)
		if c < prev {
			out[t] = c
			prev = c
		}
	}
	return out
}
