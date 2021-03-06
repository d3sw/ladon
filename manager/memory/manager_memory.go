package memory

import (
	"sort"
	"sync"

	. "github.com/d3sw/ladon"
	"github.com/pkg/errors"
)

// MemoryManager is an in-memory (non-persistent) implementation of Manager.
type MemoryManager struct {
	Policies map[string]Policy
	sync.RWMutex
}

// NewMemoryManager constructs and initializes new MemoryManager with no policies.
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		Policies: map[string]Policy{},
	}
}

// Update updates an existing policy.
func (m *MemoryManager) Update(policy Policy) error {
	m.Lock()
	defer m.Unlock()
	m.Policies[policy.GetID()] = policy
	return nil
}

// GetAll returns all policies.
func (m *MemoryManager) GetAll(limit, offset int64) (Policies, error) {
	l := len(m.Policies)
	max := limit + offset
	if max > int64(l) {
		max = int64(l)
	}

	ps := make(Policies, l)

	keys := make([]string, 0, l)
	for k := range m.Policies {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	i := 0
	for _, key := range keys {
		v, _ := m.Policies[key]
		ps[i] = v
		i++
		if i > max {
			break
		}
	}

	return ps[offset:], nil
}

// Create a new pollicy to MemoryManager.
func (m *MemoryManager) Create(policy Policy) error {
	m.Lock()
	defer m.Unlock()

	if _, found := m.Policies[policy.GetID()]; found {
		return errors.New("Policy exists")
	}

	m.Policies[policy.GetID()] = policy
	return nil
}

// Get retrieves a policy.
func (m *MemoryManager) Get(id string) (Policy, error) {
	m.RLock()
	defer m.RUnlock()
	p, ok := m.Policies[id]
	if !ok {
		return nil, errors.New("Not found")
	}

	return p, nil
}

// Delete removes a policy.
func (m *MemoryManager) Delete(id string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.Policies, id)
	return nil
}

// FindRequestCandidates returns candidates that could match the request object. It either returns
// a set that exactly matches the request, or a superset of it. If an error occurs, it returns nil and
// the error.
func (m *MemoryManager) FindRequestCandidates(r *Request) (Policies, error) {
	m.RLock()
	defer m.RUnlock()
	ps := make(Policies, len(m.Policies))
	var count int
	for _, p := range m.Policies {
		ps[count] = p
		count++
	}
	return ps, nil
}
