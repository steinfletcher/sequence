package sequence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"
)

var incTemplateFunc = &template.FuncMap{
	"inc": func(i int) int {
		return i + 1
	},
}

type (
	Diagram struct {
		title    string
		subTitle string
		name     string
		events   []interface{}
		metaJson template.JS
	}

	Request struct {
		Source      string
		Target      string
		HttpRequest http.Request
	}

	Response struct {
		Source       string
		Target       string
		HttpResponse http.Response
	}

	htmlTemplateModel struct {
		WebSequenceDSL string
		Title          string
		SubTitle       string
		BadgeClass     string
		StatusCode     string
		LogEntries     []logEntry
		MetaJSON       template.JS
	}

	logEntry struct {
		Header string
		Body   string
	}
)

func NewHttpEvents() *Diagram {
	return &Diagram{}
}

func (d Diagram) Render() (string, error) {
	statusCode, err := d.finalResponseStatus()
	if err != nil {
		return "", err
	}

	dsl, httpEventLog, err := d.transformHttpEvents(d.events)
	if err != nil {
		return "", err
	}

	htmlModel := htmlTemplateModel{
		WebSequenceDSL: dsl,
		Title:          d.title,
		SubTitle:       d.subTitle,
		BadgeClass:     getBadgeClass(statusCode),
		StatusCode:     strconv.Itoa(statusCode),
		LogEntries:     httpEventLog,
		MetaJSON:       d.metaJson,
	}
	return renderHtmlPage(htmlModel)
}

func (d *Diagram) Title(title string) *Diagram {
	d.title = title
	return d
}

func (d *Diagram) SubTitle(subTitle string) *Diagram {
	d.subTitle = subTitle
	return d
}

func (d *Diagram) Request(request Request) *Diagram {
	d.events = append(d.events, request)
	return d
}

func (d *Diagram) Response(response Response) *Diagram {
	d.events = append(d.events, response)
	return d
}

func (d *Diagram) Name(name string) *Diagram {
	d.name = name
	return d
}

func (d *Diagram) MetaJSON(jsonData template.JS) *Diagram {
	d.metaJson = jsonData
	return d
}

func (d *Diagram) finalResponseStatus() (int, error) {
	status := func(o interface{}) (int, error) {
		if res, ok := o.(Response); ok {
			return res.HttpResponse.StatusCode, nil
		}
		return -1, errors.New("final http event was not `http.Response` type")
	}
	if len(d.events) == 0 {
		return -1, nil
	}
	return status(d.events[len(d.events)-1])
}

func (d *Diagram) transformHttpEvents(httpEvents []interface{}) (string, []logEntry, error) {
	var logs []logEntry
	var dslBuffer bytes.Buffer
	for i, event := range httpEvents {
		if req, ok := event.(Request); ok {
			dslBuffer.WriteString(fmt.Sprintf("%s->%s: (%d) %s %s\n",
				req.Source,
				req.Target,
				i+1,
				req.HttpRequest.Method,
				req.HttpRequest.URL))

			entry, err := newRequestLogModel(&req.HttpRequest)
			if err != nil {
				return "", nil, err
			}
			logs = append(logs, entry)
		} else if res, ok := event.(Response); ok {
			dslBuffer.WriteString(fmt.Sprintf("%s->>%s: (%d) %d\n",
				res.Source,
				res.Target,
				i+1,
				res.HttpResponse.StatusCode))

			entry, err := newResponseLogModel(&res.HttpResponse)
			if err != nil {
				return "", nil, err
			}
			logs = append(logs, entry)
		} else {
			return "", nil, errors.New("received http event that is not a http.Request or http.Response")
		}
	}
	return dslBuffer.String(), logs, nil
}

func newRequestLogModel(req *http.Request) (logEntry, error) {
	requestDump, err := httputil.DumpRequestOut(req, false)
	if err != nil {
		return logEntry{}, err
	}
	return newEventLogEntry(requestDump, req.Body, req.Header.Get("Content-Type"))
}

func newResponseLogModel(res *http.Response) (logEntry, error) {
	requestDump, err := httputil.DumpResponse(res, false)
	if err != nil {
		return logEntry{}, err
	}
	return newEventLogEntry(requestDump, res.Body, res.Header.Get("Content-Type"))
}

func newEventLogEntry(header []byte, readCloser io.ReadCloser, contentType string) (logEntry, error) {
	if readCloser == nil {
		return logEntry{
			Header: string(header),
			Body:   "",
		}, nil
	}

	body, err := ioutil.ReadAll(readCloser)
	if err != nil {
		return logEntry{}, err
	}

	buf := new(bytes.Buffer)
	if contentType == "application/json" {
		json.Indent(buf, body, "", "    ")
	} else {
		_, err := buf.Write(body)
		if err != nil {
			panic(err)
		}
	}

	return logEntry{
		Header: string(header),
		Body:   string(buf.String()),
	}, nil
}

func getBadgeClass(statusCode int) string {
	if statusCode >= 400 && statusCode < 500 {
		return "badge badge-warning"
	} else if statusCode >= 500 {
		return "badge badge-danger"
	}
	return "badge badge-success"
}

func renderHtmlPage(model htmlTemplateModel) (string, error) {
	t, err := template.New("sequenceDiagram").
		Funcs(*incTemplateFunc).
		Parse(tmpl)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	err = t.Execute(&out, model)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

const tmpl = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.2/css/bootstrap.min.css">
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/9.12.0/styles/github.min.css" />
	<script src="https://cdnjs.cloudflare.com/ajax/libs/underscore.js/1.8.3/underscore-min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/raphael/2.2.7/raphael.min.js"></script>
    <script src="https://bramp.github.io/js-sequence-diagrams/js/sequence-diagram-min.js"></script>
    <script src="https://code.jquery.com/jquery-3.3.1.slim.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.3/umd/popper.min.js"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.1.2/js/bootstrap.min.js"></script>
</head>
<body>
<!-- THIS CODE IS AUTOGENERATED. DO NOT EDIT -->
<div class="container-fluid">
    <h1>{{ .Title }}</h1>
    <span class="{{ .BadgeClass }}">{{ .StatusCode }}</span>
    <p class="lead">{{ .SubTitle }}</p>

    <div class="card text-center">
        <div class="card-body">
            <div id="d" class="justify-content-center"></div>
        </div>
    </div>

    <br><br>
    <p class="lead">Request/Response wire representation</p>

    <table class="table">
        <thead>
        <tr>
            <th scope="col">#</th>
            <th scope="col">Payload</th>
        </tr>
        </thead>
        <tbody>
        {{ range $i, $e := .LogEntries }}

        <tr>
            <th scope="row">{{ inc $i }}</th>
            <td>
<pre>{{ $e.Header }}</pre>
{{if $e.Body }}<pre><code class="json">{{ $e.Body }}</code></pre>{{end}}
            </td>
        </tr>
        {{ end }}
        </tbody>
    </table>
</div>
<script>
    Diagram.parse("{{ .WebSequenceDSL }}").drawSVG("d", {theme: 'simple', 'font-size': 14});
</script>
<style>
    body {
        padding-top: 2rem;
        padding-bottom: 2rem;
    }
</style>
{{if $.MetaJSON }}<script type="application/json" id="metaJson">{{ $.MetaJSON }}</script>{{end}}
<script src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@9.13.1/build/highlight.min.js"></script>
<script>hljs.initHighlightingOnLoad();</script>
</body>
</html>`
