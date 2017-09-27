package ladon

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/d3sw/ladon/compiler"
)

const (
	Matchall = "all"
	Matchany = "any"
)

// BodyArrayMatchCondition is a condition which is fulfilled if the value at the
// path in the body is an array and all/any of its elements match the regex
// pattern specified in BodyArrayMatchCondition
type BodyArrayMatchCondition struct {
	Mode    string `json:"mode"`
	Path    string `json:"path"`
	Matches string `json:"matches"`
}

// Fulfills returns true if the value at the path in the body is an array and
// all/any of its elements match the regex pattern specified in BodyMatchCondition
func (c *BodyArrayMatchCondition) Fulfills(_ interface{}, r *Request) bool {
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
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		p := &DefaultPolicy{}
		reg, err := compiler.CompileRegex(c.Matches, p.GetStartDelimiter(), p.GetEndDelimiter())
		if err != nil {
			return false
		}
		v, err := JsonQuery(body, c.Path)
		return matches(v, reg, c.Mode)
	default:
		return false
	}
}

func matches(data []byte, reg *regexp.Regexp, mode string) bool {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return false
	}
	switch v.(type) {
	case []interface{}:
		for _, e := range v.([]interface{}) {
			s, ok := e.(string)
			if !ok {
				if mode == Matchall {
					return false
				}
				continue
			}
			br := reg.MatchString(s)
			if mode == Matchany && br {
				return true
			}
			if mode == Matchall && !br {
				return false
			}
		}
	}
	if mode == Matchany {
		return false
	}
	return true
}

// GetName returns the condition's name.
func (c *BodyArrayMatchCondition) GetName() string {
	return "BodyArrayMatchCondition"
}
