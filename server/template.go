package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"text/template"
)

// TemplateData holds the data available to response body templates.
type TemplateData struct {
	Query  map[string]string
	Header map[string]string
	Path   string
	Method string
}

// buildTemplateData extracts request context into a TemplateData struct.
func buildTemplateData(r *http.Request) TemplateData {
	query := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			query[k] = v[0]
		}
	}

	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return TemplateData{
		Query:  query,
		Header: headers,
		Path:   r.URL.Path,
		Method: r.Method,
	}
}

// renderTemplate attempts to render body as a Go template using request data.
// If the body contains no template directives or parsing fails, the original
// body string is returned unchanged.
func renderTemplate(body string, r *http.Request) string {
	tmpl, err := template.New("body").Parse(body)
	if err != nil {
		return body
	}

	data := buildTemplateData(r)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return body
	}
	return buf.String()
}

// renderJSONTemplate renders the body template and validates the result is
// valid JSON. Returns the rendered string and whether it is valid JSON.
func renderJSONTemplate(body string, r *http.Request) (string, bool) {
	rendered := renderTemplate(body, r)
	var js json.RawMessage
	return rendered, json.Unmarshal([]byte(rendered), &js) == nil
}

// isTemplate reports whether the given string contains any Go template
// directives (i.e. {{ ... }} action blocks).
func isTemplate(body string) bool {
	for i := 0; i < len(body)-1; i++ {
		if body[i] == '{' && body[i+1] == '{' {
			return true
		}
	}
	return false
}
