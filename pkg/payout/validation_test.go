package payout

import (
	"testing"
)

func TestValidatePayoutFrequencyValue(t *testing.T) {
	for _, freq := range []string{"THRESHOLD", "WEEKLY", "MONTHLY", "threshold", " weekly "} {
		if err := ValidatePayoutFrequencyValue(freq); err != nil {
			t.Errorf("expected %q valid, got %v", freq, err)
		}
	}
	for _, freq := range []string{"", "DAILY", "YEARLY", "thresholdx"} {
		err := ValidatePayoutFrequencyValue(freq)
		if err == nil {
			t.Fatalf("expected %q invalid", freq)
		}
		if err.Field != PayoutFieldFrequency || err.Rule != PayoutRuleInvalid {
			t.Errorf("expected frequency/invalid, got %+v", err)
		}
	}
}

func TestValidateMinimumWithdrawalValue_Boundaries(t *testing.T) {
	const sysMin = 500
	cases := []struct {
		name    string
		min     int
		wantNil bool
		field   PayoutField
		rule    PayoutRule
	}{
		{"zero", 0, false, PayoutFieldMinimum, PayoutRuleTooSmall},
		{"negative", -5, false, PayoutFieldMinimum, PayoutRuleTooSmall},
		{"one below floor", 1, false, PayoutFieldMinimum, PayoutRuleBelowSystem},
		{"at system floor", 500, true, "", ""},
		{"below system floor", 499, false, PayoutFieldMinimum, PayoutRuleBelowSystem},
		{"above floor", 501, true, "", ""},
		{"at hard cap", 100000, true, "", ""},
		{"above hard cap", 100001, false, PayoutFieldMinimum, PayoutRuleTooLarge},
	}
	if err := ValidateMinimumWithdrawalValue(1, 1); err != nil {
		t.Errorf("expected min=1 valid against floor 1, got %v", err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMinimumWithdrawalValue(tc.min, sysMin)
			if tc.wantNil {
				if err != nil {
					t.Fatalf("expected nil, got %+v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error for min=%d", tc.min)
			}
			if err.Field != tc.field || err.Rule != tc.rule {
				t.Errorf("expected %s/%s, got %+v", tc.field, tc.rule, err)
			}
		})
	}
}

func TestValidatePayoutWeekdayValue(t *testing.T) {
	for _, wd := range []int{0, 1, 6} {
		if err := ValidatePayoutWeekdayValue(wd); err != nil {
			t.Errorf("expected weekday %d valid, got %v", wd, err)
		}
	}
	for _, wd := range []int{-1, 7} {
		err := ValidatePayoutWeekdayValue(wd)
		if err == nil || err.Field != PayoutFieldWeekday || err.Rule != PayoutRuleRange {
			t.Errorf("expected weekday/range for %d, got %+v", wd, err)
		}
	}
}

func TestValidatePayoutMonthDayAndOrder(t *testing.T) {
	for _, day := range []int{1, 24, 31} {
		if err := ValidatePayoutMonthDayValue(PayoutFieldMonthStart, day); err != nil {
			t.Errorf("expected day %d valid, got %v", day, err)
		}
	}
	for _, day := range []int{0, 32} {
		err := ValidatePayoutMonthDayValue(PayoutFieldMonthEnd, day)
		if err == nil || err.Field != PayoutFieldMonthEnd || err.Rule != PayoutRuleRange {
			t.Errorf("expected month_end/range for %d, got %+v", day, err)
		}
	}
	if err := ValidatePayoutMonthOrder(24, 26); err != nil {
		t.Errorf("expected 24<=26 valid, got %v", err)
	}
	if err := ValidatePayoutMonthOrder(26, 26); err != nil {
		t.Errorf("expected equal bounds valid, got %v", err)
	}
	err := ValidatePayoutMonthOrder(27, 26)
	if err == nil || err.Rule != PayoutRuleOrder {
		t.Errorf("expected order error, got %+v", err)
	}
}
