package testing

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gophercloud/gophercloud"
	th "github.com/gophercloud/gophercloud/testhelper"
)

func buildResult(t *testing.T, resultBody, mv string) gophercloud.Result {
	var dejson interface{}
	sejson := []byte(resultBody)
	err := json.Unmarshal(sejson, &dejson)
	if err != nil {
		t.Fatal(err)
	}

	var result = gophercloud.Result{
		Body: dejson,
	}

	if mv != "" {
		result.Header = map[string][]string{}
		result.Header.Add("X-OpenStack-Nova-API-Version", mv)
	}

	return result
}

var singleResponse = `
{
	"person": {
		"name": "Bill",
		"email": "bill@example.com",
		"location": "Canada"
	}
}
`

var multiResponse = `
{
	"people": [
		{
			"name": "Bill",
			"email": "bill@example.com",
			"location": "Canada"
		},
		{
			"name": "Ted",
			"email": "ted@example.com",
			"location": "Mexico"
		}
	]
}
`

type TestPerson struct {
	Name  string `json:"-"`
	Email string `json:"email"`
}

func (r *TestPerson) UnmarshalJSON(b []byte) error {
	type tmp TestPerson
	var s struct {
		tmp
		Name string `json:"name"`
	}

	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	*r = TestPerson(s.tmp)
	r.Name = s.Name + " unmarshalled"

	return nil
}

type TestPersonExt struct {
	Location string `json:"-"`
}

func (r *TestPersonExt) UnmarshalJSON(b []byte) error {
	type tmp TestPersonExt
	var s struct {
		tmp
		Location string `json:"location"`
	}

	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	*r = TestPersonExt(s.tmp)
	r.Location = s.Location + " unmarshalled"

	return nil
}

type TestPersonWithExtensions struct {
	TestPerson    `json:"foo"`
	TestPersonExt `json:"bar"`
}

type TestPersonWithExtensionsNamed struct {
	TestPerson    TestPerson
	TestPersonExt TestPersonExt
}

// TestUnmarshalAnonymousStruct tests if UnmarshalJSON is called on each
// of the anonymous structs contained in an overarching struct.
func TestUnmarshalAnonymousStructs(t *testing.T) {
	var actual TestPersonWithExtensions
	result := buildResult(t, singleResponse, "")

	err := result.ExtractIntoStructPtr(&actual, "person")
	th.AssertNoErr(t, err)

	th.AssertEquals(t, "Bill unmarshalled", actual.Name)
	th.AssertEquals(t, "Canada unmarshalled", actual.Location)
}

// TestUnmarshalSliceofAnonymousStructs tests if UnmarshalJSON is called on each
// of the anonymous structs contained in an overarching struct slice.
func TestUnmarshalSliceOfAnonymousStructs(t *testing.T) {
	var actual []TestPersonWithExtensions
	result := buildResult(t, multiResponse, "")

	err := result.ExtractIntoSlicePtr(&actual, "people")
	th.AssertNoErr(t, err)

	th.AssertEquals(t, "Bill unmarshalled", actual[0].Name)
	th.AssertEquals(t, "Canada unmarshalled", actual[0].Location)
	th.AssertEquals(t, "Ted unmarshalled", actual[1].Name)
	th.AssertEquals(t, "Mexico unmarshalled", actual[1].Location)
}

// TestUnmarshalSliceOfStruct tests if extracting results from a "normal"
// struct still works correctly.
func TestUnmarshalSliceofStruct(t *testing.T) {
	var actual []TestPerson
	result := buildResult(t, multiResponse, "")

	err := result.ExtractIntoSlicePtr(&actual, "people")
	th.AssertNoErr(t, err)

	th.AssertEquals(t, "Bill unmarshalled", actual[0].Name)
	th.AssertEquals(t, "Ted unmarshalled", actual[1].Name)
}

// TestUnmarshalNamedStruct tests if the result is empty.
func TestUnmarshalNamedStructs(t *testing.T) {
	var actual TestPersonWithExtensionsNamed
	result := buildResult(t, singleResponse, "")

	err := result.ExtractIntoStructPtr(&actual, "person")
	th.AssertNoErr(t, err)

	th.AssertEquals(t, "", actual.TestPerson.Name)
	th.AssertEquals(t, "", actual.TestPersonExt.Location)
}

// TestUnmarshalSliceofNamedStructs tests if the result is empty.
func TestUnmarshalSliceOfNamedStructs(t *testing.T) {
	var actual []TestPersonWithExtensionsNamed
	result := buildResult(t, multiResponse, "")

	err := result.ExtractIntoSlicePtr(&actual, "people")
	th.AssertNoErr(t, err)

	th.AssertEquals(t, "", actual[0].TestPerson.Name)
	th.AssertEquals(t, "", actual[0].TestPersonExt.Location)
	th.AssertEquals(t, "", actual[1].TestPerson.Name)
	th.AssertEquals(t, "", actual[1].TestPersonExt.Location)
}

type TestServer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Enabled   bool      `json:"enabled"`
	NetworkID string    `json:"network_id" min_version:"2.10"`
	ImageID   int       `json:"image_id" max_version:"2.09"`
	FlavorID  int       `json:"flavor_id" min_version:"2.06" max_version:"2.15"`
}

var testServer205GetResponse = `
{
  "server": {
    "id": "1",
    "name": "foo",
    "tenant_id": "abcd1234",
    "user_id": "efgh5678",
    "status": "ACTIVE",
		"created_at": "2014-09-25T13:10:10Z",
    "enabled": true,
    "image_id": 12345
  }
}
`

var expected205GetResponse = map[string]interface{}{
	"id":         "1",
	"name":       "foo",
	"tenant_id":  "abcd1234",
	"user_id":    "efgh5678",
	"status":     "ACTIVE",
	"created_at": "2014-09-25T13:10:10Z",
	"enabled":    true,
	"image_id":   float64(12345),
}

var testServer210GetResponse = `
{
  "server": {
    "id": "1",
    "name": "foo",
    "tenant_id": "abcd1234",
    "user_id": "efgh5678",
    "status": "ACTIVE",
    "created_at": "2014-09-25T13:10:10Z",
    "enabled": true,
    "network_id": "aabbccdd",
    "flavor_id": 6789
  }
}
`

var expected210GetResponse = map[string]interface{}{
	"id":         "1",
	"name":       "foo",
	"tenant_id":  "abcd1234",
	"user_id":    "efgh5678",
	"status":     "ACTIVE",
	"created_at": "2014-09-25T13:10:10Z",
	"enabled":    true,
	"network_id": "aabbccdd",
	"flavor_id":  float64(6789),
}

var testServer205ListResponse = `
{
  "servers": [
    {
      "id": "1",
      "name": "foo",
      "tenant_id": "abcd1234",
      "user_id": "efgh5678",
      "status": "ACTIVE",
      "created_at": "2014-09-25T13:10:10Z",
      "enabled": true,
      "image_id": 12345
    },
    {
      "id": "2",
      "name": "bar",
      "tenant_id": "abcd1234",
      "user_id": "efgh5678",
      "status": "ACTIVE",
      "created_at": "2014-09-25T13:10:10Z",
      "enabled": false,
      "image_id": 12345
    }
  ]
}
`
var expected205ListResponse = []map[string]interface{}{
	map[string]interface{}{
		"id":         "1",
		"name":       "foo",
		"tenant_id":  "abcd1234",
		"user_id":    "efgh5678",
		"status":     "ACTIVE",
		"created_at": "2014-09-25T13:10:10Z",
		"enabled":    true,
		"image_id":   float64(12345),
	},
	map[string]interface{}{
		"id":         "2",
		"name":       "bar",
		"tenant_id":  "abcd1234",
		"user_id":    "efgh5678",
		"status":     "ACTIVE",
		"created_at": "2014-09-25T13:10:10Z",
		"enabled":    false,
		"image_id":   float64(12345),
	},
}

var testServer210ListResponse = `
{
  "servers": [
    {
      "id": "1",
      "name": "foo",
      "tenant_id": "abcd1234",
      "user_id": "efgh5678",
      "status": "ACTIVE",
      "created_at": "2014-09-25T13:10:10Z",
      "enabled": true,
      "network_id": "aabbccdd",
      "flavor_id": 6789
    },
    {
      "id": "2",
      "name": "bar",
      "tenant_id": "abcd1234",
      "user_id": "efgh5678",
      "status": "ACTIVE",
      "created_at": "2014-09-25T13:10:10Z",
      "enabled": false,
      "network_id": "aabbccdd",
      "flavor_id": 6789
    }
  ]
}
`

var expected210ListResponse = []map[string]interface{}{
	map[string]interface{}{
		"id":         "1",
		"name":       "foo",
		"tenant_id":  "abcd1234",
		"user_id":    "efgh5678",
		"status":     "ACTIVE",
		"created_at": "2014-09-25T13:10:10Z",
		"enabled":    true,
		"network_id": "aabbccdd",
		"flavor_id":  float64(6789),
	},
	map[string]interface{}{
		"id":         "2",
		"name":       "bar",
		"tenant_id":  "abcd1234",
		"user_id":    "efgh5678",
		"status":     "ACTIVE",
		"created_at": "2014-09-25T13:10:10Z",
		"enabled":    false,
		"network_id": "aabbccdd",
		"flavor_id":  float64(6789),
	},
}

func TestExtractStructIntoMap(t *testing.T) {
	var testServer TestServer

	// no microversion
	var actual map[string]interface{}
	result := buildResult(t, testServer205GetResponse, "")
	err := result.ExtractIntoStructPtr(&testServer, "server")
	th.AssertNoErr(t, err)
	err = result.ExtractIntoMapPtr(&testServer, &actual)
	th.AssertNoErr(t, err)
	th.AssertDeepEquals(t, expected205GetResponse, actual)

	// microverson 2.05
	var actual205 map[string]interface{}
	result = buildResult(t, testServer210GetResponse, "2.05")
	err = result.ExtractIntoStructPtr(&testServer, "server")
	th.AssertNoErr(t, err)
	err = result.ExtractIntoMapPtr(&testServer, &actual205)
	th.AssertNoErr(t, err)
	th.AssertDeepEquals(t, expected205GetResponse, actual205)

	// microversion 2.10
	var actual210 map[string]interface{}
	result = buildResult(t, testServer210GetResponse, "2.10")
	err = result.ExtractIntoStructPtr(&testServer, "server")
	th.AssertNoErr(t, err)
	err = result.ExtractIntoMapPtr(&testServer, &actual210)
	th.AssertNoErr(t, err)
	th.AssertDeepEquals(t, expected210GetResponse, actual210)
}

func TestExtractSliceIntoMap(t *testing.T) {
	var testServer []TestServer

	// no microversion
	var actual []map[string]interface{}
	result := buildResult(t, testServer205ListResponse, "")
	err := result.ExtractIntoSlicePtr(&testServer, "servers")
	th.AssertNoErr(t, err)
	err = result.ExtractIntoMapPtr(&testServer, &actual)
	th.AssertNoErr(t, err)
	th.AssertDeepEquals(t, expected205ListResponse, actual)

	// microversion 2.05
	var actual205 []map[string]interface{}
	result = buildResult(t, testServer205ListResponse, "2.05")
	err = result.ExtractIntoSlicePtr(&testServer, "servers")
	th.AssertNoErr(t, err)
	err = result.ExtractIntoMapPtr(&testServer, &actual205)
	th.AssertNoErr(t, err)
	th.AssertDeepEquals(t, expected205ListResponse, actual205)

	// microversion 2.10
	var actual210 []map[string]interface{}
	result = buildResult(t, testServer210ListResponse, "2.10")
	err = result.ExtractIntoSlicePtr(&testServer, "servers")
	th.AssertNoErr(t, err)
	err = result.ExtractIntoMapPtr(&testServer, &actual210)
	th.AssertNoErr(t, err)
	th.AssertDeepEquals(t, expected210ListResponse, actual210)
}

// TestUnmarshalAnonymousStructToMap tests if UnmarshalJSON is called on each
// of the anonymous structs contained in an overarching struct and is able to
// be extracted to a map.
func TestUnmarshalAnonymousStructsToMap(t *testing.T) {
	var testPerson TestPersonWithExtensions
	result := buildResult(t, singleResponse, "")

	var actual map[string]interface{}
	expected := map[string]interface{}{
		"name":     "Bill unmarshalled",
		"location": "Canada unmarshalled",
	}

	err := result.ExtractIntoStructPtr(&testPerson, "person")
	t.Logf("%#v\n", testPerson)
	th.AssertNoErr(t, err)
	err = result.ExtractIntoMapPtr(&testPerson, &actual)
	th.AssertNoErr(t, err)
	th.AssertDeepEquals(t, expected, actual)

}
