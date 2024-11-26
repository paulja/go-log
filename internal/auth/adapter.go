package auth

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
)

// Annoyingly you cannot provide a simple string for a policy, I wanted to
// be able to embed my policy for this example so I implemented the Adapter
// interface defined below so I could use a string for the policy.

/*
// Adapter is the interface for Casbin adapters.
type Adapter interface {
	// LoadPolicy loads all policy rules from the storage.
	LoadPolicy(model model.Model) error
	// SavePolicy saves all policy rules to the storage.
	SavePolicy(model model.Model) error

	// AddPolicy adds a policy rule to the storage.
	// This is part of the Auto-Save feature.
	AddPolicy(sec string, ptype string, rule []string) error
	// RemovePolicy removes a policy rule from the storage.
	// This is part of the Auto-Save feature.
	RemovePolicy(sec string, ptype string, rule []string) error
	// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
	// This is part of the Auto-Save feature.
	RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error
}
*/

type StringAdapter struct {
	policy string
}

func NewStringAdapter(policy string) *StringAdapter {
	return &StringAdapter{
		policy: policy,
	}
}

func (a *StringAdapter) LoadPolicy(m model.Model) error {
	if len(a.policy) == 0 {
		return fmt.Errorf("invalid policy, cannot be empty")
	}
	lines := strings.Split(a.policy, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		persist.LoadPolicyLine(line, m)
	}
	return nil
}

func (a *StringAdapter) SavePolicy(m model.Model) error {
	return nil
}

func (a *StringAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil
}

func (a *StringAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

func (a *StringAdapter) RemoveFilteredPolicy(
	sec string,
	ptype string,
	fieldIndex int,
	fieldValues ...string,
) error {
	return nil
}
