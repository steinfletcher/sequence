package sequence

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestCssClassBasedOnStatusCode(t *testing.T) {
	tests := []struct {
		class  string
		status int
	}{
		{"danger", 500},
		{"warning", 400},
		{"success", 200},
	}
	for _, test := range tests {
		t.Run(test.class, func(t *testing.T) {
			if class := getBadgeClass(test.status); !strings.Contains(class, test.class) {
				t.Fail()
			}
		})
	}
}

func TestComputesFinalResponseStatusCode(t *testing.T) {
	events := httpEvents()

	status, _ := events.finalResponseStatus()

	if status != http.StatusCreated {
		t.Fail()
	}
}

func TestErrorIfNoFinalResponseStatusCode(t *testing.T) {
	request, _ := http.NewRequest("POST", "http://two", strings.NewReader(`{"email": "a@b.com"}`))
	events := NewHttpEvents().
		Title("title").
		SubTitle("subTitle").
		Request(Request{
			Source:      "one",
			Target:      "two",
			HttpRequest: *request,
		})

	_, err := events.finalResponseStatus()

	if !strings.Contains(err.Error(), "was not `http.Response` type") {
		t.Fail()
	}
}

func TestSupportsRenderingOfArbitraryJsonData(t *testing.T) {
	out, err := NewHttpEvents().
		MetaJSON(`{"a": 123}`).
		Render()

	if err != nil {
		t.Fail()
	}

	if !strings.Contains(out, `<script type="application/json" id="metaJson">{"a": 123}</script>`) {
		t.Fail()
	}
}

func TestNewHttpEvents(t *testing.T) {
	expected, err := ioutil.ReadFile("testdata/expected_result.html")
	if err != nil {
		panic(err)
	}

	actual, err := httpEvents().Render()

	if err != nil {
		t.Fail()
	}
	if actual == "" {
		t.Fail()
	}
	if string(expected) != actual {
		t.Fail()
	}
}

func httpEvents() *Diagram {
	request1, _ := http.NewRequest("POST", "http://two", strings.NewReader(`{"email": "a@b.com"}`))
	request2, _ := http.NewRequest("POST", "http://three", strings.NewReader(`{"username": "a@b.com"}`))
	return NewHttpEvents().
		Title("title").
		SubTitle("subTitle").
		Request(Request{
			Source:      "one",
			Target:      "two",
			HttpRequest: *request1,
		}).
		Request(Request{
			Source:      "two",
			Target:      "three",
			HttpRequest: *request2,
		}).
		Response(Response{
			Source:       "three",
			Target:       "two",
			HttpResponse: http.Response{StatusCode: http.StatusOK},
		}).
		Response(Response{
			Source:       "two",
			Target:       "one",
			HttpResponse: http.Response{StatusCode: http.StatusCreated},
		})
}
