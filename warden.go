package ladon

import "errors"

const (
	KeyRawRequest = "http-request"
)

// Request is the warden's request object.
type Request struct {
	// Resource is the resource that access is requested to.
	Resource string `json:"resource"`

	// Action is the action that is requested on the resource.
	Action string `json:"action"`

	// Subejct is the subject that is requesting access.
	Subjects []string `json:"subject"`

	// Context is the request's environmental context.
	Context Context `json:"context"`
}

// Validate validates request is formatted correctly
func (r *Request) Validate() error {
	if r.Resource == "" {
		return errors.New("missing resource")
	}

	if r.Action == "" {
		return errors.New("missing actions")
	}

	if len(r.Subjects) == 0 {
		return errors.New("missing subjects")
	}

	return nil
}

// Warden is responsible for deciding if subject s can perform action a on resource r with context c.
type Warden interface {
	// IsAllowed returns nil if subject s can perform action a on resource r with context c or an error otherwise.
	//  if err := guard.IsAllowed(&Request{Resource: "article/1234", Action: "update", Subject: "peter"}); err != nil {
	//    return errors.New("Not allowed")
	//  }
	IsAllowed(r *Request) error
}
