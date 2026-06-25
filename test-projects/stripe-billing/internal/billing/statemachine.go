package billing

// Subscription status values.
const (
	StatusActive           = "active"
	StatusPending          = "pending"
	StatusFree             = "free"    // same as active on free plan
	StatusPastDue          = "past_due"
	StatusCancelAtPeriodEnd = "cancel_at_period_end"
	StatusCancelled        = "cancelled"
)

// validTransitions encodes all allowed from→to transitions.
// Webhook-only transitions are also listed here; CanAPIWrite governs who may
// actually drive each transition.
var validTransitions = map[string]map[string]bool{
	StatusFree: {
		StatusPending: true, // API: user subscribes to Pro
	},
	StatusActive: {
		StatusPending:           true, // API: re-subscribe after cancel
		StatusCancelAtPeriodEnd: true, // API: downgrade / soft cancel
		StatusPastDue:           true, // webhook: payment failed
		StatusCancelled:         true, // webhook: sub deleted
	},
	StatusPending: {
		StatusActive:    true, // webhook: payment succeeded
		StatusCancelled: true, // webhook: sub deleted before payment
	},
	StatusPastDue: {
		StatusActive:    true, // webhook: payment retry succeeded
		StatusCancelled: true, // webhook: sub deleted
	},
	StatusCancelAtPeriodEnd: {
		StatusCancelled: true, // webhook: period ended
		StatusActive:    true, // webhook: user re-activated before period end
	},
	StatusCancelled: {
		StatusPending: true, // API: re-subscribe
	},
}

// ValidTransition returns true if transitioning from→to is a known-valid move.
func ValidTransition(from, to string) bool {
	if from == to {
		return false // no-op transitions are not valid status changes
	}
	tos, ok := validTransitions[from]
	if !ok {
		return false
	}
	return tos[to]
}

// apiWritable is the set of statuses the API layer is permitted to write.
// All other confirmed statuses (active, past_due, cancelled) are webhook-only.
var apiWritable = map[string]bool{
	StatusPending:           true, // set when initiating a Stripe subscription
	StatusCancelAtPeriodEnd: true, // mirrors Stripe's cancel_at_period_end flag
}

// CanAPIWrite returns true if the API (not webhook) may write the given status.
func CanAPIWrite(status string) bool {
	return apiWritable[status]
}
