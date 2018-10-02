package kettei

import (
	"context"
	"errors"
)

type (
	Strategy int

	DecisionManager struct {
		voters                             []Voter
		strategy                           Strategy
		allowIfAllAbstainDecisions         bool
		allowIfEqualGrantedDeniedDecisions bool
	}
)

const (
	StrategyAffirmative = iota
	StrategyConsensus
	StrategyUnanimous
)

var (
	ErrInvalidStrategy = errors.New("invalid strategy")
)

// Decides whether the access is possible or not.
func (manager *DecisionManager) Decide(ctx *context.Context, attributes []string, subject interface{}) (bool, error) {
	switch manager.strategy {
	case StrategyAffirmative:
		return manager.decideAffirmative(ctx, attributes, subject)
	case StrategyConsensus:
		return manager.decideConsensus(ctx, attributes, subject)
	case StrategyUnanimous:
		return manager.decideUnanimous(ctx, attributes, subject)
	default:
		return false, ErrInvalidStrategy
	}
}

// Grants access if any voter returns an affirmative response.
//
// If all voters abstained from voting, the decision will be based on the allowIfAllAbstainDecisions property value
// (defaults to false).
func (manager *DecisionManager) decideAffirmative(ctx *context.Context, attributes []string, subject interface{}) (bool, error) {
	var deny int

	for _, voter := range manager.voters {
		result, err := vote(voter, ctx, attributes, subject)
		if err != nil {
			return false, err
		}

		switch result {
		case AccessGranted:
			return true, nil
		case AccessDenied:
			deny += 1
			break
		default:
			break
		}
	}

	if deny > 0 {
		return false, nil
	}

	return manager.allowIfAllAbstainDecisions, nil
}

// Grants access if there is consensus of granted against denied responses.
//
// Consensus means majority-rule (ignoring abstains) rather than unanimous agreement (ignoring abstains).
// If you require unanimity, see UnanimousBased.
//
// If there were an equal number of grant and deny votes, the decision will be based on the
// allowIfEqualGrantedDeniedDecisions property value (defaults to true).
//
// If all voters abstained from voting, the decision will be based on the allowIfAllAbstainDecisions property value
// (defaults to false).
func (manager *DecisionManager) decideConsensus(ctx *context.Context, attributes []string, subject interface{}) (bool, error) {
	var grant int
	var deny int

	for _, voter := range manager.voters {
		result, err := vote(voter, ctx, attributes, subject)
		if err != nil {
			return false, err
		}

		switch result {
		case AccessGranted:
			grant += 1
			break
		case AccessDenied:
			deny += 1
			break
		default:
			break
		}
	}

	if grant > deny {
		return true, nil
	}

	if deny > grant {
		return false, nil
	}

	if grant > 0 {
		return manager.allowIfEqualGrantedDeniedDecisions, nil
	}

	return manager.allowIfAllAbstainDecisions, nil
}

// Grants access if only grant (or abstain) votes were received.
//
// If all voters abstained from voting, the decision will be based on the allowIfAllAbstainDecisions property value
// (defaults to false).
func (manager *DecisionManager) decideUnanimous(ctx *context.Context, attributes []string, subject interface{}) (bool, error) {
	var grant int

	for _, voter := range manager.voters {
		for _, attribute := range attributes {
			result, err := vote(voter, ctx, []string{attribute}, subject)
			if err != nil {
				return false, err
			}

			switch result {
			case AccessGranted:
				grant += 1
				break
			case AccessDenied:
				return false, nil
			default:
				break
			}
		}
	}

	if grant > 0 {
		return true, nil
	}

	return manager.allowIfAllAbstainDecisions, nil
}
