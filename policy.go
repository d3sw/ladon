package ladon

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrPolicyNotFound = errors.New("policy not found")
)

// Policies is an array of policies.
type Policies []Policy

// Policy represent a policy model.
type Policy interface {
	// GetID returns the policies id.
	GetID() string

	// GetDescription returns the policies description.
	GetDescription() string

	// GetSubjects returns the policies subjects.
	GetSubjects() []string

	// AllowAccess returns true if the policy effect is allow, otherwise false.
	AllowAccess() bool

	// GetEffect returns the policies effect which might be 'allow' or 'deny'.
	GetEffect() string

	// GetResources returns the policies resources.
	GetResources() []string

	// GetActions returns the policies actions.
	GetActions() []string

	// GetConditions returns the policies conditions.
	GetConditions() Conditions

	// GetStartDelimiter returns the delimiter which identifies the beginning of a regular expression.
	GetStartDelimiter() byte

	// GetEndDelimiter returns the delimiter which identifies the end of a regular expression.
	GetEndDelimiter() byte
}

// DefaultPolicy is the default implementation of the policy interface.
type DefaultPolicy struct {
	ID          string     `json:"id" gorethink:"id"`
	Description string     `json:"description" gorethink:"description"`
	Subjects    []string   `json:"subjects" gorethink:"subjects"`
	Effect      string     `json:"effect" gorethink:"effect"`
	Resources   []string   `json:"resources" gorethink:"resources"`
	Actions     []string   `json:"actions" gorethink:"actions"`
	Conditions  Conditions `json:"conditions" gorethink:"conditions"`
}

// UnmarshalJSON overwrite own policy with values of the given in policy in JSON format
func (p *DefaultPolicy) UnmarshalJSON(data []byte) error {
	var pol = struct {
		ID          string     `json:"id" gorethink:"id"`
		Description string     `json:"description" gorethink:"description"`
		Subjects    []string   `json:"subjects" gorethink:"subjects"`
		Effect      string     `json:"effect" gorethink:"effect"`
		Resources   []string   `json:"resources" gorethink:"resources"`
		Actions     []string   `json:"actions" gorethink:"actions"`
		Conditions  Conditions `json:"conditions" gorethink:"conditions"`
	}{
		Conditions: Conditions{},
	}

	if err := json.Unmarshal(data, &pol); err != nil {
		return errors.WithStack(err)
	}

	*p = *&DefaultPolicy{
		ID:          pol.ID,
		Description: pol.Description,
		Subjects:    pol.Subjects,
		Effect:      pol.Effect,
		Resources:   pol.Resources,
		Actions:     pol.Actions,
		Conditions:  pol.Conditions,
	}
	return nil
}

// GetID returns the policies id.
func (p *DefaultPolicy) GetID() string {
	return p.ID
}

// GetDescription returns the policies description.
func (p *DefaultPolicy) GetDescription() string {
	return p.Description
}

// GetSubjects returns the policies subjects.
func (p *DefaultPolicy) GetSubjects() []string {
	return p.Subjects
}

// AllowAccess returns true if the policy effect is allow, otherwise false.
func (p *DefaultPolicy) AllowAccess() bool {
	return p.Effect == AllowAccess
}

// GetEffect returns the policies effect which might be 'allow' or 'deny'.
func (p *DefaultPolicy) GetEffect() string {
	return p.Effect
}

// GetResources returns the policies resources.
func (p *DefaultPolicy) GetResources() []string {
	return p.Resources
}

// GetActions returns the policies actions.
func (p *DefaultPolicy) GetActions() []string {
	return p.Actions
}

// GetConditions returns the policies conditions.
func (p *DefaultPolicy) GetConditions() Conditions {
	return p.Conditions
}

// GetEndDelimiter returns the delimiter which identifies the end of a regular expression.
func (p *DefaultPolicy) GetEndDelimiter() byte {
	return '>'
}

// GetStartDelimiter returns the delimiter which identifies the beginning of a regular expression.
func (p *DefaultPolicy) GetStartDelimiter() byte {
	return '<'
}

// HasIdentity returns true if the provided identity is part of the policy i.e. contained
// in the subject.
func (p *DefaultPolicy) HasIdentity(id string) bool {
	for _, sub := range p.Subjects {
		if id == sub {
			return true
		}
	}

	return false
}

// AttachIdentity adds identities to the Subjects
func (p *DefaultPolicy) AttachIdentity(ids ...string) error {
	for _, id := range ids {
		if p.HasIdentity(id) {
			return fmt.Errorf("identity already attached: %s", id)
		} else if err := checkIdentity(id); err != nil {
			return err
		}

		p.Subjects = append(p.Subjects, id)
	}
	return nil
}

func (p *DefaultPolicy) detachIdentity(id string) error {
	for i, sub := range p.Subjects {
		if id == sub {
			if len(p.Subjects) == 1 {
				p.Subjects = []string{}
			} else {
				p.Subjects = append(p.Subjects[:i], p.Subjects[i+1:]...)
			}
			return nil
		}
	}

	return fmt.Errorf("identity not attached: %s", id)
}

// DetachIdentity removes a user,group,role or any other identity from the subject of the
// policy
func (p *DefaultPolicy) DetachIdentity(ids ...string) error {
	for _, id := range ids {
		if err := p.detachIdentity(id); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates properties and values of the policy
func (p *DefaultPolicy) Validate() error {
	err := p.ValidateSubjects()
	if err == nil {
		err = p.ValidateEffect()
	}
	return err
}

// ValidateSubjects validates each subject to make sure they are syntactically correct
func (p *DefaultPolicy) ValidateSubjects() error {
	//
	// TODO:
	// for _, sub := range p.Subjects {
	// 	if parts := strings.Split(sub, ":"); len(parts) < 2 {
	// 		return fmt.Errorf("ambigous subject: %s", sub)
	// 	}
	// }

	if p.Subjects == nil || len(p.Subjects) < 1 {
		return fmt.Errorf("subject missing")
	}
	return nil
}

// ValidateEffect validates the effect
func (p *DefaultPolicy) ValidateEffect() error {
	if p.Effect != AllowAccess && p.Effect != DenyAccess {
		return fmt.Errorf("unsupported effect: %s", p.Effect)
	}
	return nil
}

func (p *DefaultPolicy) CheckID() (string, string, error) {
	id := p.GetID()
	parts := strings.Split(id, ".")
	if len(parts) < 2 {
		return "", "", errors.New("invalid id format")
	}

	return parts[0], strings.Join(parts[1:], "."), nil
}

func checkIdentity(id string) error {
	if p := strings.Split(id, ":"); len(p) < 2 {
		return fmt.Errorf("invalid id: %s", id)
	}

	return nil
}
