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
	Resources   resources        `json:"resources" gorethink:"resources"`
	Actions     actions        `json:"actions" gorethink:"actions"`
	Conditions  json.RawMessage `json:"conditions" gorethink:"conditions"`
}

type subjects struct {
	Raw      []string `json:"subjects" gorethink:"raw"`
	Compiled string   `json:"subjects" gorethink:"compiled"`
}

type resources struct {
	Raw      []string `json:"resources" gorethink:"raw"`
	Compiled string   `json:"resources" gorethink:"compiled"`
}

type actions struct {
	Raw      []string `json:"actions" gorethink:"raw"`
	Compiled string   `json:"actions" gorethink:"compiled"`
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
		Resources:   s.Resources.Raw,
		Actions:     s.Actions.Raw,
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
	s.Resources.Raw = p.GetResources()
	if err := s.compileResources(); err != nil {
		return err
	}
	s.Actions.Raw = p.GetActions()
	if err := s.compileActions(); err != nil {
		return err
	}
	s.Effect = p.GetEffect()
	s.Conditions = cs
	return nil
}

func (s *schema) compileSubject() error {
	res, err := compile(s.Subjects.Raw)
	if err != nil {
		return err
	}
	s.Subjects.Compiled = res
	return nil
}

func (s *schema) compileResources() error {
	res, err := compile(s.Resources.Raw)
	if err != nil {
		return err
	}
	s.Resources.Compiled = res
	return nil
}

func (s *schema) compileActions() error {
	res, err := compile(s.Actions.Raw)
	if err != nil {
		return err
	}
	s.Actions.Compiled = res
	return nil
}

func compile(s []string) (string, error) {
	csubs := make([]string, len(s))
	for i, s := range s {
		if cs, err := compiler.CompileRegex(s, '<', '>'); err != nil {
			return "", err
		} else {
			csubs[i] = cs.String()
		}
	}
	return strings.Join(csubs, "|"), nil
}
