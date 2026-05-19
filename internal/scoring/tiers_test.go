package scoring

import (
	"reflect"
	"testing"
)

// TestBandFor_PublishedTable verifies §3.1 against the rulebook's
// published 25-point table:
//
//	24-25 Excellent
//	22-23 Very Good
//	19-21 Good
//	18    Sufficient
//	0-17  Insufficient
func TestBandFor_PublishedTable(t *testing.T) {
	cases := []struct {
		points Points
		want   Tier
	}{
		{0, TierInsufficient},
		{10, TierInsufficient},
		{17, TierInsufficient},
		{18, TierSufficient},
		{19, TierGood},
		{20, TierGood},
		{21, TierGood},
		{22, TierVeryGood},
		{23, TierVeryGood},
		{24, TierExcellent},
		{25, TierExcellent},
	}
	for _, tc := range cases {
		got := BandFor(tc.points, 25)
		if got != tc.want {
			t.Errorf("BandFor(%d, 25) = %v, want %v", tc.points, got, tc.want)
		}
	}
}

// TestBandFor_TenPointScale exercises all five tiers on a max where
// none collapse.
func TestBandFor_TenPointScale(t *testing.T) {
	// Cutoffs at max=10: Excellent=10, VG=9, Good=8, Sufficient=7.
	cases := []struct {
		points Points
		want   Tier
	}{
		{0, TierInsufficient},
		{6, TierInsufficient},
		{7, TierSufficient},
		{8, TierGood},
		{9, TierVeryGood},
		{10, TierExcellent},
	}
	for _, tc := range cases {
		got := BandFor(tc.points, 10)
		if got != tc.want {
			t.Errorf("BandFor(%d, 10) = %v, want %v", tc.points, got, tc.want)
		}
	}
}

// TestBandFor_FivePointCollapse: per the rulebook on muzzle stability,
// Very Good and Sufficient collapse on a 5-point scale.
func TestBandFor_FivePointCollapse(t *testing.T) {
	cases := []struct {
		points Points
		want   Tier
	}{
		{0, TierInsufficient},
		{3, TierInsufficient},
		{4, TierGood},      // no Sufficient band reachable
		{5, TierExcellent}, // no Very Good band reachable
	}
	for _, tc := range cases {
		got := BandFor(tc.points, 5)
		if got != tc.want {
			t.Errorf("BandFor(%d, 5) = %v, want %v", tc.points, got, tc.want)
		}
	}
}

// TestBandFor_OverflowAndNegative documents behavior at the edges:
// over-max points stay Excellent; negative points fall to Insufficient.
func TestBandFor_OverflowAndNegative(t *testing.T) {
	if got := BandFor(100, 25); got != TierExcellent {
		t.Errorf("BandFor(100, 25) = %v, want Excellent", got)
	}
	if got := BandFor(-1, 25); got != TierInsufficient {
		t.Errorf("BandFor(-1, 25) = %v, want Insufficient", got)
	}
	if got := BandFor(0, 0); got != TierInsufficient {
		t.Errorf("BandFor(0, 0) = %v, want Insufficient (zero max)", got)
	}
}

// TestBandCutoffs_TwentyFive: full table reachable on max=25.
func TestBandCutoffs_TwentyFive(t *testing.T) {
	want := map[Tier]Points{
		TierInsufficient: 0,
		TierSufficient:   18,
		TierGood:         19,
		TierVeryGood:     22,
		TierExcellent:    24,
	}
	got := BandCutoffs(25)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BandCutoffs(25)\ngot:  %v\nwant: %v", got, want)
	}
}

// TestBandCutoffs_FivePoint: VG and Sufficient unreachable.
func TestBandCutoffs_FivePoint(t *testing.T) {
	want := map[Tier]Points{
		TierInsufficient: 0,
		TierGood:         4,
		TierExcellent:    5,
	}
	got := BandCutoffs(5)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BandCutoffs(5)\ngot:  %v\nwant: %v", got, want)
	}
}

// TestBandCutoffs_ThreePoint: only Excellent and Insufficient reachable.
func TestBandCutoffs_ThreePoint(t *testing.T) {
	want := map[Tier]Points{
		TierInsufficient: 0,
		TierExcellent:    3,
	}
	got := BandCutoffs(3)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BandCutoffs(3)\ngot:  %v\nwant: %v", got, want)
	}
}

// TestBandCutoffs_EightPoint exercises the 8-point exercise scale used
// in L1 OB (Send-Away, Retrieve on the Flat, Retrieve over Jump).
// Cutoffs: Excellent=ceil(7.68)=8, VG=ceil(6.88)=7, Good=ceil(6.08)=7
// (collapses with VG), Sufficient=ceil(5.6)=6.
func TestBandCutoffs_EightPoint(t *testing.T) {
	want := map[Tier]Points{
		TierInsufficient: 0,
		TierSufficient:   6,
		TierVeryGood:     7,
		TierExcellent:    8,
	}
	got := BandCutoffs(8)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BandCutoffs(8)\ngot:  %v\nwant: %v", got, want)
	}
}

// TestBandCutoffs_ZeroMax: degenerate case, only Insufficient.
func TestBandCutoffs_ZeroMax(t *testing.T) {
	want := map[Tier]Points{TierInsufficient: 0}
	got := BandCutoffs(0)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BandCutoffs(0)\ngot:  %v\nwant: %v", got, want)
	}
}
