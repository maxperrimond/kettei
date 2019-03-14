package kettei

import (
	"context"
	"errors"
)

type (
	Strategy int

	DecisionMakerConfig struct {
		Voters                             []Voter
		Strategy                           Strategy
		AllowIfAllAbstainDecisions         bool
		AllowIfEqualGrantedDeniedDecisions bool
	}

	DecisionMaker struct {
		voters                             []Voter
		strategy                           Strategy
		allowIfAllAbstainDecisions         bool
		allowIfEqualGrantedDeniedDecisions bool
	}

	Reason struct {
		Voter     Voter
		Attribute string
		Reason    string
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

func NewDecisionMaker(config DecisionMakerConfig) *DecisionMaker {
	return &DecisionMaker{
		voters:                             config.Voters,
		strategy:                           config.Strategy,
		allowIfAllAbstainDecisions:         config.AllowIfAllAbstainDecisions,
		allowIfEqualGrantedDeniedDecisions: config.AllowIfEqualGrantedDeniedDecisions,
	}
}

func NewDefaultDecisionMaker(voters ...Voter) *DecisionMaker {
	return NewDecisionMaker(DecisionMakerConfig{
		Voters:                     voters,
		Strategy:                   StrategyUnanimous,
		AllowIfAllAbstainDecisions: true,
	})
}

// Decides whether the access is possible or not.
func (maker *DecisionMaker) Decide(ctx context.Context, attributes []string, subject interface{}) (bool, []*Reason, error) {
	switch maker.strategy {
	case StrategyAffirmative:
		return maker.decideAffirmative(ctx, attributes, subject)
	case StrategyConsensus:
		return maker.decideConsensus(ctx, attributes, subject)
	case StrategyUnanimous:
		return maker.decideUnanimous(ctx, attributes, subject)
	default:
		return false, nil, ErrInvalidStrategy
	}
}

// Grants access if any voter returns an affirmative response.
//
// If all voters abstained from voting, the decision will be based on the allowIfAllAbstainDecisions property value
// (defaults to false).
func (maker *DecisionMaker) decideAffirmative(ctx context.Context, attributes []string, subject interface{}) (bool, []*Reason, error) {
	var deny int

	var reasons []*Reason
	for _, voter := range maker.voters {
		result, voterReasons, err := vote(voter, ctx, attributes, subject)
		if len(voterReasons) > 0 {
			reasons = append(reasons, voterReasons...)
		}

		if err != nil {
			return false, reasons, err
		}

		switch result {
		case AccessGranted:
			return true, reasons, nil
		case AccessDenied:

			deny += 1
			break
		default:
			break
		}
	}

	if deny > 0 {
		return false, reasons, nil
	}

	return maker.allowIfAllAbstainDecisions, reasons, nil
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
func (maker *DecisionMaker) decideConsensus(ctx context.Context, attributes []string, subject interface{}) (bool, []*Reason, error) {
	var grant int
	var deny int

	var reasons []*Reason
	for _, voter := range maker.voters {
		result, voterReasons, err := vote(voter, ctx, attributes, subject)
		if len(voterReasons) > 0 {
			reasons = append(reasons, voterReasons...)
		}

		if err != nil {
			return false, reasons, err
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
		return true, reasons, nil
	}

	if deny > grant {
		return false, reasons, nil
	}

	if grant > 0 {
		return maker.allowIfEqualGrantedDeniedDecisions, reasons, nil
	}

	return maker.allowIfAllAbstainDecisions, reasons, nil
}

// Grants access if only grant (or abstain) votes were received.
//
// If all voters abstained from voting, the decision will be based on the allowIfAllAbstainDecisions property value
// (defaults to false).
func (maker *DecisionMaker) decideUnanimous(ctx context.Context, attributes []string, subject interface{}) (bool, []*Reason, error) {
	var grant int

	var reasons []*Reason
	for _, voter := range maker.voters {
		for _, attribute := range attributes {
			result, voterReasons, err := vote(voter, ctx, []string{attribute}, subject)
			if len(voterReasons) > 0 {
				reasons = append(reasons, voterReasons...)
			}

			if err != nil {
				return false, reasons, err
			}

			switch result {
			case AccessGranted:
				grant += 1
				break
			case AccessDenied:
				return false, reasons, nil
			default:
				break
			}
		}
	}

	if grant > 0 {
		return true, reasons, nil
	}

	return maker.allowIfAllAbstainDecisions, reasons, nil
}
