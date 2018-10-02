# kettei

Symfony [voters](https://symfony.com/doc/current/security/voters.html) in Golang

# Example

```go
package main

import (
	"context"
	
	"github.com/maxperrimond/kettei"
)

// Some subject struct
type Pancake struct{}

// Declare some custom voter
type PancakeVoter struct{}

func (voter *PancakeVoter) Support(attribute string, subject interface{}) bool {
	_, ok := subject.(*Pancake)

	return attribute == "eat" && ok
}

func (voter *PancakeVoter) VoteOnAttribute(ctx context.Context, attribute string, subject interface{}) (bool, error) {
	pancake := subject.(*Pancake)
	
	switch attribute {
	case "eat":
	  return ctx.Value("authenticated") == true, nil
	default:
	   return false, nil
	}
}

func main() {
	// Declare our decision maker
	maker := kettei.NewDecisionMaker(kettei.DecisionMakerConfig{
        Strategy: kettei.StrategyUnanimous,
        Voters: []kettei.Voter{
            &PancakeVoter{},
	    },
	})
	// You can also use NewDefaultDecisionMaker to get default unanimous DecisionMaker for most common cases

	// Somewhere in your app ...
	ok, _ := maker.Decide(ctx, []string{"eat"}, pancake)
	if !ok {
		SendErrorToUser()
	}
}
```

## Installation

    go get github.com/maxperrimond/kettei
