package kettei

import "context"

type (
	Access int

	Voter interface {
		Support(attribute string, subject interface{}) bool
		VoteOnAttribute(ctx context.Context, attribute string, subject interface{}) (bool, error)
	}
)

const (
	AccessGranted Access = 1
	AccessAbstain Access = 0
	AccessDenied  Access = -1
)

func vote(voter Voter, ctx context.Context, attributes []string, subject interface{}) (Access, error) {
	var vote = AccessAbstain

	for _, attribute := range attributes {
		if !voter.Support(attribute, subject) {
			continue
		}

		vote = AccessDenied

		result, err := voter.VoteOnAttribute(ctx, attribute, subject)
		if err != nil {
			return vote, err
		}

		if result {
			vote = AccessGranted
		}
	}

	return vote, nil
}
