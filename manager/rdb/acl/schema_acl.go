package acl

import (
	"encoding/json"
	. "github.com/d3sw/ladon"
	"github.com/pkg/errors"
)

type schema struct {
	ID          string          `json:"id" gorethink:"id"`
	Description string          `json:"description" gorethink:"description"`
	Subject     string          `json:"subjects" gorethink:"subjects"`
	Effect      string          `json:"effect" gorethink:"effect"`
	Resource    string          `json:"resources" gorethink:"resources"`
	Action      string          `json:"actions" gorethink:"actions"`
	Conditions  json.RawMessage `json:"conditions" gorethink:"conditions"`
}

func (s *schema) getPolicy() (*DefaultPolicy, error) {
	if s == nil {
		return nil, nil
	}
	cs := Conditions{}
	if err := cs.UnmarshalJSON(s.Conditions); err != nil {
		return nil, errors.WithStack(err)
	}
	return &DefaultPolicy{
		ID:          s.ID,
		Description: s.Description,
		Subjects:    []string{s.Subject},
		Effect:      s.Effect,
		Resources:   []string{s.Resource},
		Actions:     []string{s.Action},
		Conditions:  cs,
	}, nil
}

func (s *schema) getSchema(p Policy) error {
	cs, err := p.GetConditions().MarshalJSON()
	if err != nil {
		return err
	}
	s.ID = p.GetID()
	s.Description = p.GetDescription()

	sbjs := p.GetSubjects()
	if len(sbjs) > 0 {
		s.Subject = sbjs[0]
	}

	ress := p.GetSubjects()
	if len(ress) > 0 {
		s.Resource = ress[0]
	}

	acts := p.GetSubjects()
	if len(acts) > 0 {
		s.Action = acts[0]
	}

	s.Effect = p.GetEffect()
	s.Conditions = cs
	return nil
}
