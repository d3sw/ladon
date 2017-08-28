package rdb

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/d3sw/ladon"
	r "gopkg.in/gorethink/gorethink.v3"
)

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
	m := NewRdbManager(session, "policies")

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

	fmt.Printf("compiled subject is %s\n", s.Subjects.Compiled)
	tdata := []schema{s}
	res, err := r.Expr(tdata).Filter(func(item r.Term) r.Term {
		//return r.Expr("aaa").Match(r.Expr("(?i)").Add(item.Field("subjects").Field("compiled")))
		return r.Expr("xyz").Match(item.Field("subjects").Field("compiled")) //case sensitive
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
