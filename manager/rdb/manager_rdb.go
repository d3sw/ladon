package rdb

import (
	"fmt"

	. "github.com/d3sw/ladon"
	"github.com/pkg/errors"
	r "gopkg.in/gorethink/gorethink.v3"
)

// RdbManager is a rethinkdb implementation of Manager to store policies persistently.
type RdbManager struct {
	session *r.Session
	table   r.Term
	s       SchemaManager
}

// NewRdbManager initializes a new RdbManager for given session.
func NewRdbManager(session *r.Session, table string, s SchemaManager) *RdbManager {
	return &RdbManager{
		session: session,
		table:   r.Table(table),
		s:       s,
	}
}

// Create inserts a new policy.
func (m *RdbManager) Create(policy Policy) error {
	s := m.s.NewSchema()
	s.PopulateWithPolicy(policy)
	if _, err := m.table.Insert(s).RunWrite(m.session); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Update updates an existing policy.
func (m *RdbManager) Update(policy Policy) error {
	s := m.s.NewSchema()
	s.PopulateWithPolicy(policy)
	if _, err := m.table.Get(s.GetID()).Update(s).RunWrite(m.session); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Get retrieves a policy.
func (m *RdbManager) Get(id string) (Policy, error) {
	res, err := m.table.Get(id).Run(m.session)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Close()

	for v := range m.s.ProcessResult(res) {
		if v.Err != nil {
			return nil, v.Err
		}

		p, err := v.Schema.GetPolicy()
		if err != nil {
			return nil, err
		}

		return p, err
	}

	return nil, fmt.Errorf("failed to find policy %s", id)
}

// Delete removes a policy.
func (m *RdbManager) Delete(id string) error {
	if _, err := m.table.Get(id).Delete().RunWrite(m.session); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetAll returns all policies.
func (m *RdbManager) GetAll(limit, offset int64) (Policies, error) {
	res, err := m.table.Skip(offset).Limit(limit).Run(m.session)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Close()

	var policies Policies
	for s := range m.s.ProcessResult(res) {
		if s.Err != nil {
			return nil, s.Err
		}
		p, err := s.Schema.GetPolicy()
		if err != nil {
			return policies, errors.WithStack(err)
		}
		policies = append(policies, p)
	}

	if err := res.Err(); err != nil {
		return nil, err
	}
	return policies, nil
}

type FilterFunc func(t r.Term) r.Term

// FindRequestCandidates returns candidates that could match the request object. It either returns
// a set that exactly matches the request, or a superset of it. If an error occurs, it returns nil and
// the error.
func (m *RdbManager) FindRequestCandidates(req *Request) (Policies, error) {
	mp := map[string]bool{}
	var policies Policies
	if err := req.Validate(); err != nil {
		return nil, errors.WithStack(err)
	}

	for _, s := range req.Subjects {
		res, err := m.s.GetRequestCandidatesTerm(m.table, s, req.Resource, req.Action).Run(m.session)
		if err != nil {
			return nil, err
		}

		for v := range m.s.ProcessResult(res) {
			if v.Err != nil {
				return nil, errors.WithStack(v.Err)
			}

			if _, ok := mp[v.Schema.GetID()]; !ok {
				p, err := v.Schema.GetPolicy()
				if err != nil {
					return nil, errors.WithStack(err)
				}
				mp[v.Schema.GetID()] = true
				policies = append(policies, p)
			}
		}

		// This call isn't deferred to prevent leaks in loop
		res.Close()
	}
	return policies, nil
}
