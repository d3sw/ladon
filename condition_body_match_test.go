package ladon

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonBodyMatch(t *testing.T) {
	for _, c := range []struct {
		matches string
		path    string
		pass    bool
	}{
		{matches: "roles:xxx-yyy.<.*>", path: ".subjects[0]", pass: true},
		{matches: "[\"roles:xxx-yyy.<[^,]+>\"<(,\"roles:xxx-yyy[.]{1}[^,]+\")+>]", path: ".subjects", pass: true},
		{matches: "[\"roles:xxx-yyy.<[^,]+>\"<(,\"roles:xxx-yy[.]{1}[^,]+\")+>]", path: ".subjects", pass: false},
		{matches: "<[\"roles:xxx-yyy[.]{1}[^,]+\"(,\"roles:xxx-yyy[.]{1}[^,]+\")*>]", path: ".subjects", pass: true},
		{matches: "<[\"roles:xxx-yyy[.]{1}[^,]+\"(,\"roles:xxx-yy[.]{1}[^,]+\")*>]", path: ".subjects", pass: false},
		{matches: "allow", path: ".effect", pass: true},
		{matches: "deny", path: ".effect", pass: false},
	} {
		condition := &BodyMatchCondition{
			Path:    c.path,
			Matches: c.matches,
		}
		var body = `
		{
	      "subjects":["roles:xxx-yyy.admin","roles:xxx-yyy.read"],
	      "actions":["<.*>"],
	      "resources":["zzz:<.*>"],
	      "effect":"allow"
	  }
`
		r, _ := http.NewRequest("POST", "http://fuac.xxx.xxx.xxx/v1/app/test", bytes.NewBuffer([]byte(body)))
		r.Header.Set("Content-type", "application/json")

		lr := new(Request)
		lr.Context = Context{}
		lr.Context[KeyRawRequest] = r
		assert.Equal(t, c.pass, condition.Fulfills(nil, lr), "%s", c.matches)
	}
}
