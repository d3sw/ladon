package rdb

import (
	"sync"

	. "github.com/d3sw/ladon"
	"github.com/pkg/errors"
	r "gopkg.in/gorethink/gorethink.v3"
)

// RdbManager is a rethinkdb implementation of Manager to store policies persistently.
type RdbManager struct {
	sync.RWMutex
	session  *r.Session
	table    r.Term
	policies map[string]Policy
}

// NewRdbManager initializes a new RdbManager for given session.
func NewRdbManager(session *r.Session, table string) *RdbManager {
	return &RdbManager{
		session:  session,
		table:    r.Table(table),
		policies: map[string]Policy{},
	}
}

// Init initializes RdbManager.
func (m *RdbManager) Init() (error, <-chan error) {
	if err := m.loadPolicies(); err != nil {
		return errors.WithStack(err), nil
	}
	errch := make(chan error)
	go m.getChangeFeeds(errch)
	return nil, errch
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
	m.RLock()
	defer m.RUnlock()
	p, ok := m.policies[id]
	if !ok {
		return nil, errors.New("Not found")
	}

	return p, nil
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
	ps := make(Policies, len(m.policies))
	i := 0
	for _, p := range m.policies {
		ps[i] = p
		i++
	}

	if offset+limit > int64(len(m.policies)) {
		limit = int64(len(m.policies))
		offset = 0
	}

	return ps[offset:limit], nil
}

// FindRequestCandidates returns candidates that could match the request object. It either returns
// a set that exactly matches the request, or a superset of it. If an error occurs, it returns nil and
// the error.
func (m *RdbManager) FindRequestCandidates(r *Request) (Policies, error) {
	m.RLock()
	defer m.RUnlock()
	ps := make(Policies, len(m.policies))
	count := 0
	for _, p := range m.policies {
		ps[count] = p
		count++
	}
	return ps, nil
}

func (m *RdbManager) loadPolicies() error {
	res, err := m.table.Run(m.session)
	if err != nil {
		return errors.WithStack(err)
	}

	var s schema
	m.Lock()
	defer m.Unlock()
	for res.Next(&s) {
		p, err := s.getPolicy()
		if err != nil {
			return errors.WithStack(err)
		}
		m.policies[p.GetID()] = p
	}
	return nil
}

func (m *RdbManager) getChangeFeeds(errch chan error) {
	res, err := m.table.Changes().Run(m.session)
	if err != nil {
		errch <- errors.WithStack(err)
	}
	defer res.Close()

	var ms map[string]*schema
	for res.Next(&ms) {
		m.Lock()
		if ms["new_val"] == nil {
			if ms["old_val"] != nil {
				delete(m.policies, ms["old_val"].ID)
			}
		} else {
			if ms["old_val"] != nil && ms["new_val"].ID != ms["old_val"].ID {
				delete(m.policies, ms["old_val"].ID)
			}
			p, err := ms["new_val"].getPolicy()
			if err != nil {
				errch <- errors.WithStack(err)
			}
			m.policies[p.GetID()] = p
		}
		m.Unlock()
	}

	if res.Err() != nil {
		errch <- errors.WithStack(res.Err())
	}
	errch <- nil
}
