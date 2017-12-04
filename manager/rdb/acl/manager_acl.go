package acl

import (
	"fmt"

	. "github.com/d3sw/ladon"
	"github.com/pkg/errors"
	r "gopkg.in/gorethink/gorethink.v3"
)

// ACLManager is a rethinkdb implementation of Manager to store policies persistently.
type ACLManager struct {
	session *r.Session
	table   r.Term
}

// NewACLManager initializes a new ACLManager for given session.
func NewACLManager(session *r.Session, table string) *ACLManager {
	return &ACLManager{
		session: session,
		table:   r.Table(table),
	}
}

// Create inserts a new policy.
func (m *ACLManager) Create(policy Policy) error {
	s := &schema{}
	s.getSchema(policy)
	if _, err := m.table.Insert(s).RunWrite(m.session); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Update updates an existing policy.
func (m *ACLManager) Update(policy Policy) error {
	s := &schema{}
	s.getSchema(policy)
	if _, err := m.table.Get(s.ID).Update(s).RunWrite(m.session); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Get retrieves a policy.
func (m *ACLManager) Get(id string) (Policy, error) {
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
func (m *ACLManager) Delete(id string) error {
	if _, err := m.table.Get(id).Delete().RunWrite(m.session); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetAll returns all policies.
func (m *ACLManager) GetAll(limit, offset int64) (Policies, error) {
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
func (m *ACLManager) FindRequestCandidates(req *Request) (Policies, error) {
	mp := map[string]bool{}
	var policies Policies
	//if err := req.Validate(); err != nil {
	//	return nil, err
	//}

	for _, s := range req.Subjects {
		filterResultCh := m.filter(func(t r.Term) r.Term {
			tr := t.Field("subject").Match(s).
				And(t.Field("resource").Match(req.Resource)).
				And(t.Field("action").Match(req.Action))

			return tr
		})

		for fr := range filterResultCh {
			if fr.err != nil {
				return nil, errors.WithStack(fr.err)
			}

			if _, ok := mp[fr.s.ID]; !ok {
				p, err := fr.s.getPolicy()
				if err != nil {
					return nil, errors.WithStack(err)
				}
				mp[fr.s.ID] = true
				policies = append(policies, p)
			}
		}
	}
	return policies, nil
}

type filterFunc func(t r.Term) r.Term

type filterResult struct {
	s   schema
	err error
}

func (m *ACLManager) filter(f filterFunc) chan *filterResult {
	ch := make(chan *filterResult)

	go func() {
		defer close(ch)
		res, err := m.table.Filter(f).Run(m.session)
		if err != nil {
			ch <- &filterResult{err: err}
			return
		}
		defer res.Close()

		var s schema
		for res.Next(&s) {
			ch <- &filterResult{
				s: s,
			}
		}
	}()

	return ch
}
