package ladon

import (
	"net/http"

	"github.com/SermoDigital/jose/jwt"
	"github.com/d3sw/fuac/foken"
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

// Warden is responsible for deciding if subject s can perform action a on resource r with context c.
type Warden interface {
	// IsAllowed returns nil if subject s can perform action a on resource r with context c or an error otherwise.
	//  if err := guard.IsAllowed(&Request{Resource: "article/1234", Action: "update", Subject: "peter"}); err != nil {
	//    return errors.New("Not allowed")
	//  }
	IsAllowed(r *Request) error
}

// BuildPolicyRequest builds a PolicyRequest from the http.Request.  It sets all fields
// in claims, http headers, and http query pararms to the context
func BuildRequest(r *http.Request, claims jwt.Claims) *Request {
	req := &Request{
		Subjects: []string{},
		Action:   r.Method,
	}

	ctx := Context{"remoteIP": r.RemoteAddr}

	// Set claims to context
	if claims != nil {
		for k, v := range claims {
			ctx[k] = v
		}
	}

	// Add user:email as subject
	eml := claims.Get(foken.ClaimEmail)
	if email, ok := eml.(string); ok {
		req.Subjects = append(req.Subjects, "user:"+email)
		delete(ctx, foken.ClaimEmail)
	}
	// Add user:username as subject
	puser := claims.Get(foken.ClaimUsername)
	if username, ok := puser.(string); ok {
		req.Subjects = append(req.Subjects, "user:"+username)
		delete(ctx, foken.ClaimUsername)
	}

	//
	// TODO: add role subjects
	//

	// Set headers as context
	for k := range r.Header {
		ctx[k] = r.Header.Get(k)
	}
	delete(ctx, "Authorization")

	// Set query params as context
	vals := r.URL.Query()
	for k, v := range vals {
		ctx[k] = v
	}

	req.Context = ctx

	return req
}
