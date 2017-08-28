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
}

// NewRdbManager initializes a new RdbManager for given session.
func NewRdbManager(session *r.Session, table string) *RdbManager {
	return &RdbManager{
		session: session,
		table:   r.Table(table),
	}
}

// Create inserts a new policy.
func (m *RdbManager) Create(policy Policy) error {
	s := &schema{}
	s.getSchema(policy)
	if _, err := m.table.Insert(s).RunWrite(m.session); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Update updates an existing policy.
func (m *RdbManager) Update(policy Policy) error {
	s := &schema{}
	s.getSchema(policy)
	if _, err := m.table.Get(s.ID).Update(s).RunWrite(m.session); err != nil {
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
	var s schema
	for res.Next(&s) {
		p, err := s.getPolicy()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return p, nil
	}
	return nil, fmt.Errorf("Failed to find policy %s", id)
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
	var s schema
	for res.Next(&s) {
		p, err := s.getPolicy()
		if err != nil {
			return policies, errors.WithStack(err)
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// FindRequestCandidates returns candidates that could match the request object. It either returns
// a set that exactly matches the request, or a superset of it. If an error occurs, it returns nil and
// the error.
func (m *RdbManager) FindRequestCandidates(req *Request) (Policies, error) {
	mp := map[string]bool{}
	var policies Policies
	for _, s := range req.Subjects {
		res, err := m.table.Filter(func(t r.Term) r.Term {
			return r.Expr(s).Match(t.Field("subjects").Field("compiled"))
		}).Run(m.session)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer res.Close()
		var s schema
		for res.Next(&s) {
			if _, ok := mp[s.ID]; !ok {
				p, err := s.getPolicy()
				if err != nil {
					return nil, errors.WithStack(err)
				}
				mp[s.ID] = true
				policies = append(policies, p)
			}
		}
	}
	return policies, nil
}
