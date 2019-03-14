package kettei

import "context"

type (
	Access int

	Voter interface {
		Support(attribute string, subject interface{}) bool
		VoteOnAttribute(ctx context.Context, attribute string, subject interface{}) (bool, string, error)
	}
)

const (
	AccessGranted Access = 1
	AccessAbstain Access = 0
	AccessDenied  Access = -1
)

func vote(voter Voter, ctx context.Context, attributes []string, subject interface{}) (Access, []*Reason, error) {
	var vote = AccessAbstain

	var reasons []*Reason
	for _, attribute := range attributes {
		if !voter.Support(attribute, subject) {
			continue
		}

		vote = AccessDenied

		result, reasonMessage, err := voter.VoteOnAttribute(ctx, attribute, subject)
		if reasonMessage != "" {
			reason := Reason{
				Reason:    reasonMessage,
				Attribute: attribute,
				Voter:     voter,
			}

			reasons = append(reasons, &reason)
		}

		if err != nil {
			return vote, reasons, err
		}

		if result {
			vote = AccessGranted
		}
	}

	return vote, reasons, nil
}
