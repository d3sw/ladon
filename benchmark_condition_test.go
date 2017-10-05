package ladon_test

import (
	"bytes"
	"github.com/d3sw/ladon"
	"net/http"
	"testing"
)

func benchmarkLadonConditions(b *testing.B, cs []ladon.Condition) {
	b.ResetTimer()

	var body = `{"subjects":["role:deluxe-dummy.admin"], "actions":["<.*>"], "resources":["dummy:<.*>"], "effect":"allow"}`

	r, _ := http.NewRequest("POST", "http://fuac.xxx.xxx.xxx/v1/app/test", bytes.NewBuffer([]byte(body)))
	r.Header.Set("Content-type", "application/json")

	lr := new(ladon.Request)
	lr.Context = ladon.Context{}
	lr.Context[ladon.KeyRawRequest] = r

	for n := 0; n < b.N; n++ {
		for _, c := range cs {
			c.Fulfills("test", lr)
		}
	}
}

func Benchmark_String(b *testing.B) {
	c := []ladon.Condition{&ladon.StringMatchCondition{Matches: "test"}}

	benchmarkLadonConditions(b, c)
}

func Benchmark_Body(b *testing.B) {
	c := []ladon.Condition{
		&ladon.BodyMatchCondition{Path: ".subjects", Matches: "role:deluxe-dummy.<.+>"},
	}

	benchmarkLadonConditions(b, c)
}

func Benchmark_BodyArray(b *testing.B) {
	c := []ladon.Condition{
		&ladon.BodyArrayMatchCondition{Mode: "all", Path: ".subjects", Matches: "role:deluxe-dummy.<.+>"},
	}

	benchmarkLadonConditions(b, c)
}

func Benchmark_String_Body(b *testing.B) {
	c := []ladon.Condition{
		&ladon.StringMatchCondition{Matches: "test"},
		&ladon.BodyMatchCondition{Path: ".subjects", Matches: "role:deluxe-dummy.<.+>"},
	}

	benchmarkLadonConditions(b, c)
}

func Benchmark_String_BodyArray(b *testing.B) {
	c := []ladon.Condition{
		&ladon.StringMatchCondition{Matches: "test"},
		&ladon.BodyArrayMatchCondition{Mode: "all", Path: ".subjects", Matches: "role:deluxe-dummy.<.+>"},
	}

	benchmarkLadonConditions(b, c)
}

func Benchmark_Body_BodyArray(b *testing.B) {
	c := []ladon.Condition{
		&ladon.BodyMatchCondition{Path: ".subjects", Matches: "role:deluxe-dummy.<.+>"},
		&ladon.BodyArrayMatchCondition{Mode: "all", Path: ".subjects", Matches: "role:deluxe-dummy.<.+>"},
	}

	benchmarkLadonConditions(b, c)
}

func Benchmark_String_Body_BodyArray(b *testing.B) {
	c := []ladon.Condition{
		&ladon.StringMatchCondition{Matches: "test"},
		&ladon.BodyMatchCondition{Path: ".subjects", Matches: "role:deluxe-dummy.<.+>"},
		&ladon.BodyArrayMatchCondition{Mode: "all", Path: ".subjects", Matches: "role:deluxe-dummy.<.+>"},
	}

	benchmarkLadonConditions(b, c)
}
