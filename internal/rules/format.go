package rules

import "strings"

// SuggestFormat recommends a workable competition format when the configured
// style is incompatible with the actual entrant count or pool sizes.
func SuggestFormat(style string, entrants, pools int) string {
	if entrants < 2 {
		return "no_contest"
	}
	lower := strings.ToLower(style)
	if pools <= 0 {
		return "round_robin"
	}
	if strings.Contains(lower, "elimination") {
		if entrants < 6 {
			return "round_robin"
		}
		if entrants/pools < 4 {
			return "round_robin"
		}
	}
	return style
}

// ThirdPlacePlan avoids repeat pairings: if the two bronze contenders already
// met, two bronze medals are awarded instead of a rematch.
func ThirdPlacePlan(haveMet bool, allowRematch bool) string {
	if !allowRematch || haveMet {
		return "dual_bronze"
	}
	return "rematch"
}

// CanAccept reports whether another entrant fits under a capacity limit.
// A capacity of zero means unlimited.
func CanAccept(capacity, current int) bool {
	return capacity <= 0 || current < capacity
}

// CapRemaining returns available places, or -1 when unlimited.
func CapRemaining(capacity, current int) int {
	if capacity <= 0 {
		return -1
	}
	remaining := capacity - current
	if remaining < 0 {
		return 0
	}
	return remaining
}
