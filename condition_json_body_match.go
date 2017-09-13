package ladon

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/d3sw/ladon/compiler"
)

// JsonBodyMatchCondition is a condition which is fulfilled if the value at the
// path in the json body matches the regex pattern specified in JsonBodyMatchCondition
type JsonBodyMatchCondition struct {
	Path    string `json:"path"`
	Matches string `json:"matches"`
}

// Fulfills returns true if the value at the path in the json body matches
// the regex pattern specified in JsonBodyMatchCondition
func (c *JsonBodyMatchCondition) Fulfills(_ interface{}, r *Request) bool {
	req, ok := r.Context[KeyRawRequest].(*http.Request)
	if !ok {
		return false
	}
	contentType := req.Header.Get("Content-type")
	if !strings.EqualFold(contentType, "application/json") {
		return false
	}
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
}

// GetName returns the condition's name.
func (c *JsonBodyMatchCondition) GetName() string {
	return "JsonBodyMatchCondition"
}
