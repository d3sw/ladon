package ladon

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonBodyArrayMatch(t *testing.T) {
	for _, c := range []struct {
		mode    string
		matches string
		path    string
		pass    bool
	}{
		{mode: "any", matches: "roles:xxx-yyy.<.*>", path: ".subjects[0]", pass: false},
		{mode: "any", matches: "roles:xxx-yyy.<.*>", path: ".subjects", pass: true},
		{mode: "all", matches: "roles:xxx-yyy.<.*>", path: ".subjects", pass: false},
		{mode: "all", matches: "roles:xxx-yy<.*>.<.*>", path: ".subjects", pass: true},
	} {
		condition := &BodyArrayMatchCondition{
			Mode:    c.mode,
			Path:    c.path,
			Matches: c.matches,
		}
		var body = `
		{
	      "subjects":["roles:xxx-yyy.admin","roles:xxx-yy.read"],
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
