package ladon

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/d3sw/ladon/compiler"
)

// BodyMatchCondition is a condition which is fulfilled if the value at the
// path in the body matches the regex pattern specified in BodyMatchCondition
type BodyMatchCondition struct {
	Path    string `json:"path"`
	Matches string `json:"matches"`
}

// Fulfills returns true if the value at the path in the body matches the regex
// pattern specified in BodyMatchCondition
func (c *BodyMatchCondition) Fulfills(_ interface{}, r *Request) bool {
	req, ok := r.Context[KeyRawRequest].(*http.Request)
	if !ok {
		return false
	}
	contentType := strings.ToLower(req.Header.Get("Content-type"))
	switch contentType {
	case "application/json":
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return false
		}
		s, err := String(JsonQuery(body, c.Path))
		if err != nil {
			return false
		}
		p := &DefaultPolicy{}
		reg, err := compiler.CompileRegex(c.Matches, p.GetStartDelimiter(), p.GetEndDelimiter())
		if err != nil {
			return false
		}
		return reg.MatchString(s)
	default:
		return false
	}
}

// GetName returns the condition's name.
func (c *BodyMatchCondition) GetName() string {
	return "BodyMatchCondition"
}
