package billing

// Plan name constants.
const (
	PlanFree       = "free"
	PlanPro        = "pro"
	PlanEnterprise = "enterprise"

	FreePlanUserLimit = 5
	ProPlanUserLimit  = 50
)

// PlanLimits describes the resource limits for a plan.
type PlanLimits struct {
	Plan         string
	UserLimit    int
	APICallLimit *int // nil for Free (no quota enforcement)
}

// GetPlanLimits returns the PlanLimits for the given plan name.
// apiQuotaPro is the configurable API call limit for the Pro plan.
func GetPlanLimits(plan string, apiQuotaPro int) PlanLimits {
	switch plan {
	case PlanPro:
		q := apiQuotaPro
		return PlanLimits{
			Plan:         PlanPro,
			UserLimit:    ProPlanUserLimit,
			APICallLimit: &q,
		}
	case PlanEnterprise:
		// Enterprise limits are managed out-of-band; no self-serve quota.
		return PlanLimits{
			Plan:      PlanEnterprise,
			UserLimit: 0, // unlimited / custom
		}
	default: // free
		return PlanLimits{
			Plan:         PlanFree,
			UserLimit:    FreePlanUserLimit,
			APICallLimit: nil,
		}
	}
}
