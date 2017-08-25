package rdb

import (
	"encoding/json"
	"strings"

	. "github.com/d3sw/ladon"
	"github.com/d3sw/ladon/compiler"
	"github.com/pkg/errors"
)

type schema struct {
	ID          string          `json:"id" gorethink:"id"`
	Description string          `json:"description" gorethink:"description"`
	Subjects    subjects        `json:"subjects" gorethink:"subjects"`
	Effect      string          `json:"effect" gorethink:"effect"`
	Resources   []string        `json:"resources" gorethink:"resources"`
	Actions     []string        `json:"actions" gorethink:"actions"`
	Conditions  json.RawMessage `json:"conditions" gorethink:"conditions"`
}

type subjects struct {
	Raw      []string `json:"subjects" gorethink:"raw"`
	Compiled string   `json:"subjects" gorethink:"compiled"`
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
		Subjects:    s.Subjects.Raw,
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
	s.Subjects.Raw = p.GetSubjects()
	if err := s.compileSubject(); err != nil {
		return err
	}
	s.Effect = p.GetEffect()
	s.Resources = p.GetResources()
	s.Actions = p.GetActions()
	s.Conditions = cs
	return nil
}

func (s *schema) compileSubject() error {
	csubs := make([]string, len(s.Subjects.Raw))
	for i, s := range s.Subjects.Raw {
		if cs, err := compiler.CompileRegex(s, '<', '>'); err != nil {
			return err
		} else {
			csubs[i] = cs.String()
		}
	}
	s.Subjects.Compiled = strings.Join(csubs, "|")
	return nil
}
