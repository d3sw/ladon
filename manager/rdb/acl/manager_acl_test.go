package acl

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/d3sw/ladon"
	r "gopkg.in/gorethink/gorethink.v3"
)

func Benchmark_FindRequestCandidates(b *testing.B) {
	b.Skip("Benchmark_FindRequestCandidates should only be enabled with a running fuac db")
	var session *r.Session
	copts := r.ConnectOpts{
		Address: "fuac-db-http.service.owf-dev:28015",
	}
	session, err := r.Connect(copts)
	if err != nil {
		b.Fatal(err)
	}
	m := NewACLManager(session, "acl_policies")

	req := &Request{
		Subjects: []string{"fuac"},
		Resource: "/policy/",
		Action:   "POST",
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, err := m.FindRequestCandidates(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Test_FindRequestCandidates(t *testing.T) {
	t.Skip("Test_FindRequestCandidates should only be enabled with a running fuac db")
	var session *r.Session
	copts := r.ConnectOpts{
		Address: "fuac-db-http.service.owf-dev:28015",
	}
	session, err := r.Connect(copts)
	if err != nil {
		t.Fatal(err)
	}
	m := NewACLManager(session, "acl_policies")

	req := &Request{
		Subjects: []string{"root"},
	}
	cands, err := m.FindRequestCandidates(req)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cands {
		fmt.Printf("user:test is included by c.ID: %s / subjects: %s\n", c.GetID(), strings.Join(c.GetSubjects(), ","))
	}
}

func Test_FindRequestCandidates_Complex(t *testing.T) {
	t.Skip("Test_FindRequestCandidates_Complex should only be enabled with a running fuac db")
	var session *r.Session
	copts := r.ConnectOpts{
		Address: "fuac-db-http.service.owf-dev:28015",
	}
	session, err := r.Connect(copts)
	if err != nil {
		t.Fatal(err)
	}
	m := NewACLManager(session, "acl_policies")

	req := &Request{
		Subjects: []string{"fuac"},
		Resource: "fuac:<.*>/policy/<.*>",
		Action:   "POST",
	}

	cands, err := m.FindRequestCandidates(req)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cands {
		fmt.Printf("user:test is included by c.ID: %s / subjects: %s; actions: %s; resources: %s\n",
			c.GetID(),
			strings.Join(c.GetSubjects(), ","),
			strings.Join(c.GetActions(), ","),
			strings.Join(c.GetResources(), ","),
		)
	}
}

func Test_FilterPolicies(t *testing.T) {
	var session *r.Session
	copts := r.ConnectOpts{
		Address: "fuac-db-http.service.owf-dev:28015",
	}
	session, err := r.Connect(copts)
	if err != nil {
		t.Fatal(err)
	}

	p := DefaultPolicy{
		ID:          "123",
		Description: "e",
		Subjects:    []string{"<.+>"},
		Effect:      "allow",
		Resources:   []string{"rsc1"},
		Actions:     []string{"GET"},
		Conditions:  Conditions{},
	}
	var s schema
	err = s.getSchema(&p)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("subject is %s\n", s.Subject)
	tdata := []schema{s}
	res, err := r.Expr(tdata).Filter(func(item r.Term) r.Term {
		//return r.Expr("aaa").Match(r.Expr("(?i)").Add(item.Field("subjects").Field("compiled")))
		return item.Field("subject").Match("xyz") //case sensitive
	}).Run(session)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Close()
	var s_ schema
	if err = res.One(&s_); err != nil {
		t.Fatal(err)
	}
	if s_.ID != s.ID {
		t.Fatal("Data mismtach")
	}
}
