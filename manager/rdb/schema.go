package rdb

import (
	"encoding/json"

	. "github.com/d3sw/ladon"

	"github.com/pkg/errors"
)

type schema struct {
	ID          string          `json:"id" gorethink:"id"`
	Description string          `json:"description" gorethink:"description"`
	Subjects    []string        `json:"subjects" gorethink:"subjects"`
	Effect      string          `json:"effect" gorethink:"effect"`
	Resources   []string        `json:"resources" gorethink:"resources"`
	Actions     []string        `json:"actions" gorethink:"actions"`
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
		Subjects:    s.Subjects,
		Effect:      s.Effect,
		Resources:   s.Resources,
		Actions:     s.Actions,
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
	s.Subjects = p.GetSubjects()
	s.Effect = p.GetEffect()
	s.Resources = p.GetResources()
	s.Actions = p.GetActions()
	s.Conditions = cs
	return nil
}
