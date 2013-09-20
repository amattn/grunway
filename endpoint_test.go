package grunway

import (
	"fmt"
	"net/url"
	"testing"
)

func TestParsePathValid(t *testing.T) {

	inputs := []string{
		"http://host/api/v1/entity/",
		"http://host/api/v1/entity/action/",

		"/api/v1/entity/",
		"/api/v2/entity/",
		"/api/v3/entity/",

		"/api/v1/entity/action",
		"/api/v1/entity/action/",
		"/api/v1/entity/123",
		"/api/v1/entity/123/action",
		"/api/v1/entity/123/action/extra1",

		"/api/v1/entity/123/action/extra1/?a=b&c=d",
		"/api/v1/entity/123/action/?a=b&c=d",
		"/api/v1/entity/123/?a=b&c=d",
		"/api/v1/entity/?a=b&c=d",
		"/api/v1/entity/?a=b&c=d",
	}
	expecteds := []Endpoint{
		Endpoint{"1", "entity", 0, "", []string{}, 0, nil},
		Endpoint{"1", "entity", 0, "action", []string{"action"}, 0, nil},

		Endpoint{"1", "entity", 0, "", []string{}, 0, nil},
		Endpoint{"2", "entity", 0, "", []string{}, 0, nil},
		Endpoint{"3", "entity", 0, "", []string{}, 0, nil},

		Endpoint{"1", "entity", 0, "action", []string{"action"}, 0, nil},
		Endpoint{"1", "entity", 0, "action", []string{"action"}, 0, nil},
		Endpoint{"1", "entity", 123, "", []string{"123"}, 0, nil},
		Endpoint{"1", "entity", 123, "action", []string{"123", "action"}, 0, nil},
		Endpoint{"1", "entity", 123, "action", []string{"123", "action", "extra1"}, 0, nil},

		Endpoint{"1", "entity", 123, "action", []string{"123", "action", "extra1"}, 0, nil},
		Endpoint{"1", "entity", 123, "action", []string{"123", "action"}, 0, nil},
		Endpoint{"1", "entity", 123, "", []string{"123"}, 0, nil},
		Endpoint{"1", "entity", 0, "", []string{}, 0, nil},
		Endpoint{"1", "entity", 0, "", []string{}, 0, nil},
	}

	// sanity check
	if len(inputs) != len(expecteds) {
		t.Fatalf("likely error in test, len(inputs) != len(expecteds")
	}

	for i := 0; i < len(inputs); i++ {
		if doesMatch, reason := inputMatchesExpected(inputs[i], expecteds[i]); doesMatch == false {
			t.Errorf("index:%d, input does not match expected: %s", i, reason)
		}
	}
}
func inputMatchesExpected(input string, expected Endpoint) (gotExpected bool, reason string) {

	parsedURL, err := url.Parse(input)
	if err != nil {
		return false, fmt.Sprintln("url.Parse returned error, ", err)
	}

	candidate, clientErr, serverErr := parsePath(parsedURL, "/api/")

	if clientErr != nil {
		return false, fmt.Sprintln("parsePath returned clientErr, ", clientErr)
	}
	if serverErr != nil {
		return false, fmt.Sprintln("parsePath returned serverErr, ", serverErr)
	}

	if candidate.isEqual(&expected) == false {
		return false, fmt.Sprintln("endpoint candidate DNE expected\ncandidate:", candidate, "\n expected:", expected)
	}

	return true, ""
}

func (endpointPtr *Endpoint) isEqual(otherPtr *Endpoint) bool {

	// helper
	stringSlicesAreEqual := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		} else {
			for i := 0; i < len(a); i++ {
				if a[i] != b[i] {
					return false
				}
			}
		}
		return true
	}

	if endpointPtr.VersionStr != otherPtr.VersionStr {
		return false
	}
	if endpointPtr.Version() != otherPtr.Version() {
		return false
	}
	if endpointPtr.EntityName != otherPtr.EntityName {
		return false
	}
	if endpointPtr.PrimaryKey != otherPtr.PrimaryKey {
		return false
	}

	// extras
	if stringSlicesAreEqual(endpointPtr.Extras, otherPtr.Extras) == false {
		return false
	}

	return true
}
