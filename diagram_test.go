package sequence

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewHttpEvents(t *testing.T) {
	request1, _ := http.NewRequest("POST", "http://two", strings.NewReader(`{"email": "a@b.com"}`))
	request2, _ := http.NewRequest("POST", "http://three", strings.NewReader(`{"username": "a@b.com"}`))
	htmlOutput, err := NewHttpEvents().
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
		}).
		Render()

	if err != nil {
		t.Fail()
	}

	if htmlOutput == "" {
		t.Fail()
	}

	f, err := os.Create("idx.html")
	if err != nil {
		panic(err)
	}
	f.WriteString(htmlOutput)
	s, _ := filepath.Abs("idx.html")
	fmt.Printf("file://%s\n", s)
}
