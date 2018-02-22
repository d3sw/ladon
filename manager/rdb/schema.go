package rdb

import (
	"encoding/json"
	"strings"

	. "github.com/d3sw/ladon"
	"github.com/d3sw/ladon/compiler"
	"github.com/pkg/errors"
	r "gopkg.in/gorethink/gorethink.v3"
)

type ProcessResult struct {
	Schema Schema
	Err    error
}

type DBResult interface {
	Next(dest interface{}) bool
	Err() error
}

type SchemaManager interface {
	NewSchema() Schema
	GetFilterFunc(subject, resource, action string) FilterFunc
	GetRequestCandidatesTerm(table r.Term, subject, resource, action string) r.Term
	ProcessResult(r DBResult) chan *ProcessResult
}

type Schema interface {
	GetID() string
	GetPolicy() (Policy, error)
	PopulateWithPolicy(Policy) error
}

// TODO: Everything what is below to move outside to place where it's used

type PolicySchema struct {
	ID          string          `json:"id" gorethink:"id"`
	Description string          `json:"description" gorethink:"description"`
	Subjects    subjects        `json:"subjects" gorethink:"subjects"`
	Effect      string          `json:"effect" gorethink:"effect"`
	Resources   resources       `json:"resources" gorethink:"resources"`
	Actions     actions         `json:"actions" gorethink:"actions"`
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

func (s *PolicySchema) GetID() string {
	return s.ID
}

func (s *PolicySchema) GetPolicy() (Policy, error) {
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

func (s *PolicySchema) PopulateWithPolicy(p Policy) error {
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

func (s *PolicySchema) compileSubject() error {
	res, err := compile(s.Subjects.Raw)
	if err != nil {
		return err
	}
	s.Subjects.Compiled = res
	return nil
}

func (s *PolicySchema) compileResources() error {
	res, err := compile(s.Resources.Raw)
	if err != nil {
		return err
	}
	s.Resources.Compiled = res
	return nil
}

func (s *PolicySchema) compileActions() error {
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

type PolicySchemaManager struct {}

func (_ *PolicySchemaManager) NewSchema() Schema {
	return &PolicySchema{}
}

func (_ *PolicySchemaManager) GetFilterFunc(sbj, res, act string) FilterFunc {
	return func(t r.Term) r.Term {
		tr := r.Expr(sbj).Match(t.Field("subjects").Field("compiled")).
			And(
				r.Expr(res).Match(t.Field("resources").Field("compiled")),
			).
			And(
				r.Expr(act).Match(t.Field("actions").Field("compiled")),
			)

		return tr
	}
}

// GetRequestCandidates term returns term to provide getRequestCandidates ability
func (sm *PolicySchemaManager) GetRequestCandidatesTerm(table r.Term, sbj, res, act string) r.Term {
	return table.Filter(sm.GetFilterFunc(sbj, res, act))
}

func (_ *PolicySchemaManager) ProcessResult(r DBResult) chan *ProcessResult {
	schemaCh := make(chan *ProcessResult)

	go func() {
		var sc *PolicySchema
		defer close(schemaCh)

		for r.Next(&sc) {
			schemaCh <- &ProcessResult{
				Schema: sc,
				Err:    r.Err(),
			}
		}

		if err := r.Err(); err != nil {
			schemaCh <- &ProcessResult{
				Schema: nil,
				Err:    r.Err(),
			}
		}
	}()

	return schemaCh
}
