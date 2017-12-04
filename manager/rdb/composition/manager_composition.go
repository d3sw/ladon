package composition

import (
	. "github.com/d3sw/ladon"
	"github.com/d3sw/ladon/manager/rdb"
	"github.com/d3sw/ladon/manager/rdb/acl"
	"errors"
)

type CompositionManager struct {
	a acl.ACLManager
	r rdb.RdbManager
}

func NewManager(a acl.ACLManager, r rdb.RdbManager) *CompositionManager {
	return &CompositionManager{
		a: a,
		r: r,
	}
}

// Create inserts a new policy.
func (m *CompositionManager) Create(policy Policy) error {
	return errors.New("composition manager cannot create policies")
}

// Update updates an existing policy.
func (m *CompositionManager) Update(policy Policy) error {
	return errors.New("composition manager cannot update policies")
}

// Get retrieves a policy.
func (m *CompositionManager) Get(id string) (Policy, error) {
	return nil, errors.New("composition manager cannot get policies")
}

// Delete removes a policy.
func (m *CompositionManager) Delete(id string) error {
	return errors.New("composition manager cannot delete policies")
}

// GetAll returns all policies.
func (m *CompositionManager) GetAll(limit, offset int64) (Policies, error) {
	return nil, errors.New("composition manager cannot return you all policies")
}

// FindRequestCandidates returns candidates that could match the request object. It either returns
// a set that exactly matches the request, or a superset of it. If an error occurs, it returns nil and
// the error.
func (m *CompositionManager) FindRequestCandidates(req *Request) (Policies, error) {
	aclP, err := m.a.FindRequestCandidates(req)
	if err != nil {
		return nil, err
	}

	rdbP, err := m.r.FindRequestCandidates(req)
	if err != nil {
		return nil, err
	}

	return append(aclP, rdbP...), nil
}
