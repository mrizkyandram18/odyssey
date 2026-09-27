package payout

import (
	"fmt"
)

// PayoutField identifies which per-user payout field failed validation.
type PayoutField string

const (
	PayoutFieldFrequency  PayoutField = "payout_frequency"
	PayoutFieldMinimum    PayoutField = "minimum_withdrawal_coins"
	PayoutFieldWeekday    PayoutField = "payout_weekday"
	PayoutFieldMonthStart PayoutField = "payout_month_start_day"
	PayoutFieldMonthEnd   PayoutField = "payout_month_end_day"
)

// PayoutRule identifies the violated rule for a field.
type PayoutRule string

const (
	// PayoutRuleInvalid is an unknown frequency value.
	PayoutRuleInvalid PayoutRule = "invalid"
	// PayoutRuleTooSmall is a numeric value below its hard lower bound.
	PayoutRuleTooSmall PayoutRule = "too_small"
	// PayoutRuleTooLarge is a numeric value above its hard upper bound.
	PayoutRuleTooLarge PayoutRule = "too_large"
	// PayoutRuleBelowSystem is a minimum below the system floor.
	PayoutRuleBelowSystem PayoutRule = "below_system_minimum"
	// PayoutRuleRange is a schedule value outside its valid range.
	PayoutRuleRange PayoutRule = "range"
	// PayoutRuleOrder is a month window with start after end.
	PayoutRuleOrder PayoutRule = "order"
)

// PayoutValidationError is returned by the canonical per-user payout
// validators below. It carries no HTTP semantics: handlers map it to
// user-facing messages with shared.WriteJSONError and keep ownership of
// status codes and response shapes.
type PayoutValidationError struct {
	Field PayoutField
	Rule  PayoutRule
}

func (e *PayoutValidationError) Error() string {
	return fmt.Sprintf("invalid payout config: %s (%s)", string(e.Field), string(e.Rule))
}

// ValidatePayoutFrequencyValue reports whether a provided frequency value
// is a known frequency. Empty/unknown values are invalid; callers decide
// whether the field is required by only invoking this when provided
// (partial-update path) or always (complete-schedule path).
func ValidatePayoutFrequencyValue(freq string) *PayoutValidationError {
	if !IsValidFrequency(freq) {
		return &PayoutValidationError{Field: PayoutFieldFrequency, Rule: PayoutRuleInvalid}
	}
	return nil
}

// ValidateMinimumWithdrawalValue enforces the hard bounds 1..100000 and
// the system floor, in that order. Bounds and floor are the pre-existing
// business rules, centralized here from the admin handlers.
// Precondition: sysMin <= 100000, enforced by the only system-minimum writer
// (HandleUpdateSystemConfig caps 1..100000). Under it, this order reproduces
// every existing handler's accept/reject outcome and message selection.
func ValidateMinimumWithdrawalValue(min, sysMin int) *PayoutValidationError {
	if min < 1 {
		return &PayoutValidationError{Field: PayoutFieldMinimum, Rule: PayoutRuleTooSmall}
	}
	if min < sysMin {
		return &PayoutValidationError{Field: PayoutFieldMinimum, Rule: PayoutRuleBelowSystem}
	}
	if min > 100000 {
		return &PayoutValidationError{Field: PayoutFieldMinimum, Rule: PayoutRuleTooLarge}
	}
	return nil
}

// ValidatePayoutWeekdayValue enforces weekday 0..6 (0=Sunday).
func ValidatePayoutWeekdayValue(weekday int) *PayoutValidationError {
	if weekday < 0 || weekday > 6 {
		return &PayoutValidationError{Field: PayoutFieldWeekday, Rule: PayoutRuleRange}
	}
	return nil
}

// ValidatePayoutMonthDayValue enforces month day 1..31 for the given field.
func ValidatePayoutMonthDayValue(field PayoutField, day int) *PayoutValidationError {
	if day < 1 || day > 31 {
		return &PayoutValidationError{Field: field, Rule: PayoutRuleRange}
	}
	return nil
}

// ValidatePayoutMonthOrder enforces month window start <= end.
func ValidatePayoutMonthOrder(start, end int) *PayoutValidationError {
	if start > end {
		return &PayoutValidationError{Field: PayoutFieldMonthStart, Rule: PayoutRuleOrder}
	}
	return nil
}
