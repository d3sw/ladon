# d3sw/Ladon
d3sw/Ladon is based on the [ory/Ladon](https://github.com/ory/ladon) which is a open source library for access management similar to [AWS IAM Policies](http://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html).

## Differences from ory/Ladon
* It includes a storage adapter for [RethinkDB](https://rethinkdb.com).
* It considers multiple representatives, i.e. username, email, of a subject in the authorization process.

---
## Services using d3sw/Ladon library
* [Fuac](https://github.com/d3sw/fuac): a Universal Access Control system meant for any type of application

---

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Installation](#installation)
- [Concepts](#concepts)
- [Usage](#usage)
  - [Policies](#policies)
    - [Conditions](#conditions)
      - [CIDR Condition](#cidr-condition)
      - [String Equal Condition](#string-equal-condition)
      - [Subject Condition](#subject-condition)
      - [String Pairs Equal Condition](#string-pairs-equal-condition)
      - [Adding Custom Conditions](#adding-custom-conditions)
    - [Persistence](#persistence)
  - [Access Control (Warden)](#access-control-warden)
- [Examples](#examples)
- [Good to know](#good-to-know)
- [Useful commands](#useful-commands)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

Ladon utilizes ory-am/dockertest for tests.
Please refer to [ory-am/dockertest](https://github.com/ory-am/dockertest) for more information of how to setup testing environment.

## Installation

```
go get github.com/d3sw/ladon
```

We recommend to use [Glide](https://github.com/Masterminds/glide) for dependency management. Ladon uses [semantic
versioning](http://semver.org/) and versions beginning with zero (`0.1.2`) might introduce backwards compatibility
breaks with [each minor version](http://semver.org/#how-should-i-deal-with-revisions-in-the-0yz-initial-development-phase).

## Concepts

Ladon is an access control library that answers the question:

> **Who** is **able** to do **what** on **something** given some **context**

* **Who**: An arbitrary unique subject name, for example "ken" or "printer-service.mydomain.com".
* **Able**: The effect which can be either "allow" or "deny".
* **What**: An arbitrary action name, for example "delete", "create" or "scoped:action:something".
* **Something**: An arbitrary unique resource name, for example "something", "resources.articles.1234" or some uniform
    resource name like "urn:isbn:3827370191".
* **Context**: The current context containing information about the environment such as the IP Address,
    request date, the resource owner name, the department ken is working in or any other information you want to pass along.
    (optional)

To decide what the answer is, Ladon uses policy documents which can be represented as JSON

```json
{
  "description": "One policy to rule them all.",
  "subjects": ["users:<[peter|ken]>", "users:maria", "groups:admins"],
  "actions" : ["delete", "<[create|update]>"],
  "effect": "allow",
  "resources": [
    "resources:articles:<.*>",
    "resources:printer"
  ],
  "conditions": {
    "remoteIP": {
        "type": "CIDRCondition",
        "options": {
            "cidr": "192.168.0.1/16"
        }
    }
  }
}
```

and can answer access requests that look like:

```json
{
  "subject": "users:peter",
  "action" : "delete",
  "resource": "resource:articles:ladon-introduction",
  "context": {
    "remoteIP": "192.168.0.5"
  }
}
```

However, Ladon does not come with a HTTP or server implementation. It does not restrict JSON either. We believe that it is your job to decide
if you want to use Protobuf, RESTful, HTTP, AMPQ, or some other protocol. It's up to you to write server!

The following example should give you an idea what a RESTful flow *could* look like. Initially we create a policy by
POSTing it to an artificial HTTP endpoint:

```
> curl \
      -X POST \
      -H "Content-Type: application/json" \
      -d@- \
      "https://my-ladon-implementation.localhost/policies" <<EOF
        {
          "description": "One policy to rule them all.",
          "subjects": ["users:<[peter|ken]>", "users:maria", "groups:admins"],
          "actions" : ["delete", "<[create|update]>"],
          "effect": "allow",
          "resources": [
            "resources:articles:<.*>",
            "resources:printer"
          ],
          "conditions": {
            "remoteIP": {
                "type": "CIDRCondition",
                "options": {
                    "cidr": "192.168.0.1/16"
                }
            }
          }
        }
  EOF
```

Then we test if "peter" (ip: "192.168.0.5") is allowed to "delete" the "ladon-introduction" article:

```
> curl \
      -X POST \
      -H "Content-Type: application/json" \
      -d@- \
      "https://my-ladon-implementation.localhost/warden" <<EOF
        {
          "subject": "users:peter",
          "action" : "delete",
          "resource": "resource:articles:ladon-introduction",
          "context": {
            "remoteIP": "192.168.0.5"
          }
        }
  EOF

{
    "allowed": true
}
```

## Usage

We already discussed two essential parts of Ladon: policies and access control requests. Let's take a closer look at those two.

### Policies

Policies are the basis for access control decisions. Think of them as a set of rules. In this library, policies
are abstracted as the `ladon.Policy` interface, and Ladon comes with a standard implementation of this interface
which is `ladon.DefaultPolicy`. Creating such a policy could look like:

```go
import "github.com/d3sw/ladon"

var pol = &ladon.DefaultPolicy{
	// A required unique identifier. Used primarily for database retrieval.
	ID: "68819e5a-738b-41ec-b03c-b58a1b19d043",

	// A optional human readable description.
	Description: "something humanly readable",

	// A subject can be an user or a service. It is the "who" in "who is allowed to do what on something".
	// As you can see here, you can use regular expressions inside < >.
	Subjects: []string{"max", "peter", "<zac|ken>"},

	// Which resources this policy affects.
	// Again, you can put regular expressions in inside < >.
	Resources: []string{"myrn:some.domain.com:resource:123", "myrn:some.domain.com:resource:345", "myrn:something:foo:<.+>"},

	// Which actions this policy affects. Supports RegExp
	// Again, you can put regular expressions in inside < >.
	Actions: []string{"<create|delete>", "get"},

	// Should access be allowed or denied?
	// Note: If multiple policies match an access request, ladon.DenyAccess will always override ladon.AllowAccess
	// and thus deny access.
	Effect: ladon.AllowAccess,

	// Under which conditions this policy is "active".
	Conditions: ladon.Conditions{
		// In this example, the policy is only "active" when the requested subject is the owner of the resource as well.
		"resourceOwner": &ladon.EqualsSubjectCondition{},

		// Additionally, the policy will only match if the requests remote ip address matches address range 127.0.0.1/32
		"remoteIPAddress": &ladon.CIDRCondition{
			CIDR: "127.0.0.1/32",
		},
	},
}
```

#### Conditions

Conditions are functions returning true or false given a context. Because conditions implement logic, they must
be programmed. Adding conditions to a policy consist of two parts, a key name and an implementation of `ladon.Condition`:

```go
// StringEqualCondition is an exemplary condition.
type StringEqualCondition struct {
	Equals string `json:"equals"`
}

// Fulfills returns true if the given value is a string and is the
// same as in StringEqualCondition.Equals
func (c *StringEqualCondition) Fulfills(value interface{}, _ *ladon.Request) bool {
	s, ok := value.(string)

	return ok && s == c.Equals
}

// GetName returns the condition's name.
func (c *StringEqualCondition) GetName() string {
	return "StringEqualCondition"
}

var pol = &ladon.DefaultPolicy{
    // ...
    Conditions: ladon.Conditions{
        "some-arbitrary-key": &StringEqualCondition{
            Equals: "the-value-should-be-this"
        }
    },
}
```

The default implementation of `Policy` supports JSON un-/marshalling. In JSON, this policy would look like:

```json
{
  "conditions": {
    "some-arbitrary-key": {
        "type": "StringEqualCondition",
        "options": {
            "equals": "the-value-should-be-this"
        }
    }
  }
}
```

As you can see, `type` is the value that `StringEqualCondition.GetName()` is returning and `options` is used to
set the value of `StringEqualCondition.Equals`.

This condition is fulfilled by (we will cover the warden in the next section)

```go
var err = warden.IsAllowed(&ladon.Request{
    // ...
    Context: &ladon.Context{
        "some-arbitrary-key": "the-value-should-be-this",
    },
}
```

but not by

```go
var err = warden.IsAllowed(&ladon.Request{
    // ...
    Context: &ladon.Context{
        "some-arbitrary-key": "some other value",
    },
}
```

and neither by:

```go
var err = warden.IsAllowed(&ladon.Request{
    // ...
    Context: &ladon.Context{
        "same value but other key": "the-value-should-be-this",
    },
}
```

Ladon ships with a couple of default conditions:

##### [CIDR Condition](condition_cidr.go)

The CIDR condition matches CIDR IP Ranges. Using this condition would look like this in JSON:

```json
{
    "conditions": {
        "remoteIP": {
            "type": "CIDRCondition",
            "options": {
                "cidr": "192.168.0.1/16"
            }
        }
    }
}
```

and in Go:

```go
var pol = &ladon.DefaultPolicy{
    Conditions: ladon.Conditions{
        "remoteIPAddress": &ladon.CIDRCondition{
            CIDR: "192.168.0.1/16",
        },
    },
}
```

In this case, we expect that the context of an access request contains a field `"remoteIpAddress"` matching
the CIDR `"192.168.0.1/16"`, for example `"192.168.0.5"`.


##### [String Equal Condition](condition_string_equal.go)

Checks if the value passed in the access request's context is identical with the string that was given initially

```go
var pol = &ladon.DefaultPolicy{
    Conditions: ladon.Conditions{
        "some-arbitrary-key": &ladon.StringEqualCondition{
            Equals: "the-value-should-be-this"
        }
    },
}
```

and would match in the following case:

```go
var err = warden.IsAllowed(&ladon.Request{
    // ...
    Context: &ladon.Context{
         "some-arbitrary-key": "the-value-should-be-this",
    },
}
```

##### [Subject Condition](condition_subject_equal.go)

Checks if the access request's subject is identical with the string that was given initially

```go
var pol = &ladon.DefaultPolicy{
    Conditions: ladon.Conditions{
        "some-arbitrary-key": &ladon.EqualsSubjectCondition{}
    },
}
```

and would match

```go
var err = warden.IsAllowed(&ladon.Request{
    // ...
    Subject: "peter",
    Context: &ladon.Context{
         "some-arbitrary-key": "peter",
    },
}
```

but not:

```go
var err = warden.IsAllowed(&ladon.Request{
    // ...
    Subject: "peter",
    Context: &ladon.Context{
         "some-arbitrary-key": "max",
    },
}
```

##### [String Pairs Equal Condition](condition_string_pairs_equal.go)

Checks if the value passed in the access request's context contains two-element arrays
and that both elements in each pair are equal.

```go
var pol = &ladon.DefaultPolicy{
    Conditions: ladon.Conditions{
        "some-arbitrary-key": &ladon.StringPairsEqualCondition{}
    },
}
```

and would match

```go
var err = warden.IsAllowed(&ladon.Request{
    // ...
    Context: &ladon.Context{
         "some-arbitrary-key": [
             ["some-arbitrary-pair-value", "some-arbitrary-pair-value"],
             ["some-other-arbitrary-pair-value", "some-other-arbitrary-pair-value"],
         ]
    },
}
```

but not:

```go
var err = warden.IsAllowed(&ladon.Request{
    // ...
    Context: &ladon.Context{
         "some-arbitrary-key": [
             ["some-arbitrary-pair-value", "some-other-arbitrary-pair-value"],
         ]
    },
}
```

##### Adding Custom Conditions

You can add custom conditions by appending it to `ladon.ConditionFactories`:

```go
import "github.com/d3sw/ladon"

func main() {
    // ...

    ladon.ConditionFactories[new(CustomCondition).GetName()] = func() Condition {
        return new(CustomCondition)
    }

    // ...
}
```

#### Persistence

d3sw/Ladon implements interface `ladon.Manager` for RethinkDB besides the In-Memory option originally provided by ory/Ladon.

**In-Memory**

```go
import (
	"github.com/d3sw/ladon"
	manager "github.com/d3sw/ladon/manager/memory"
)


func main() {
	warden := &ladon.Ladon{
		Manager: manager.NewMemoryManager(),
	}
	err := warden.Manager.Create(pol)

    // ...
}
```

**RethinkDB**

```go
import "github.com/d3sw/ladon"
import manager "github.com/d3sw/ladon/manager/rdb"
import r "gopkg.in/gorethink/gorethink.v3"

func main() {
    var session *r.Session
	opts := r.ConnectOpts{
		Address: "127.0.0.1:28015",
	}
	session, e := r.Connect(opts)
	if e != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

    warden := ladon.Ladon{
        Manager: manager.NewRdbManager(session, "policies"),
    }

    // ...
}
```

### Access Control (Warden)

Now that we have defined our policies, we can use the warden to check if a request is valid.
`ladon.Ladon`, which is the default implementation for the `ladon.Warden` interface defines `ladon.Ladon.IsAllowed()` which
will return `nil` if the access request can be granted and an error otherwise.

```go
import "github.com/d3sw/ladon"

func main() {
    // ...

    if err := warden.IsAllowed(&ladon.Request{
        Subject: "peter",
        Action: "delete",
        Resource: "myrn:some.domain.com:resource:123",
        Context: ladon.Context{
            "ip": "127.0.0.1",
        },
    }); err != nil {
        log.Fatal("Access denied")
    }

    // ...
}
```

## Examples

Check out [ladon_test.go](ladon_test.go) which includes a couple of policies and tests cases. You can run the code with `go test -run=TestLadon -v .`

## Good to know

* All checks are *case sensitive* because subject values could be case sensitive IDs.
* If `ladon.Ladon` is not able to match a policy with the request, it will default to denying the request and return an error.

Ladon does not use reflection for matching conditions to their appropriate structs due to security considerations.

## Useful commands

**Create mocks**
```sh
mockgen -package ladon_test -destination manager_mock_test.go github.com/d3sw/ladon Manager
```
