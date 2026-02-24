package frontend

import (
	"bufio"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type EndpointDoc struct {
	Name        string
	Method      string // GET, POST, DELETE
	MethodClass string // get, post, delete — for CSS
	URL         string
	Auth        string
	Description string
}

type EndpointGroup struct {
	Title     string
	IsData    bool // true for "Data Endpoints" — controls blue accent
	Endpoints []EndpointDoc
}

// parseEndpointGroups reads README.md and extracts endpoint sections into structured data.
func parseEndpointGroups(path string) ([]EndpointGroup, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var groups []EndpointGroup
	var currentGroup *EndpointGroup
	var currentEndpoint *EndpointDoc

	flushEndpoint := func() {
		if currentEndpoint != nil && currentGroup != nil {
			currentGroup.Endpoints = append(currentGroup.Endpoints, *currentEndpoint)
			currentEndpoint = nil
		}
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "## "):
			flushEndpoint()
			if currentGroup != nil {
				groups = append(groups, *currentGroup)
			}
			if strings.Contains(line, "Endpoints") {
				title := strings.TrimPrefix(line, "## ")
				currentGroup = &EndpointGroup{
					Title:  title,
					IsData: strings.Contains(title, "Data"),
				}
			} else {
				currentGroup = nil
			}

		case strings.HasPrefix(line, "### ") && currentGroup != nil:
			flushEndpoint()
			name := strings.TrimPrefix(line, "### ")
			currentEndpoint = &EndpointDoc{Name: name}

		case strings.HasPrefix(line, "- **URL**:") && currentEndpoint != nil:
			currentEndpoint.URL = extractValue(line)

		case strings.HasPrefix(line, "- **Method**:") && currentEndpoint != nil:
			currentEndpoint.Method = extractValue(line)
			currentEndpoint.MethodClass = strings.ToLower(currentEndpoint.Method)

		case strings.HasPrefix(line, "- **Auth**:") && currentEndpoint != nil:
			currentEndpoint.Auth = extractValue(line)

		case strings.HasPrefix(line, "- **Description**:") && currentEndpoint != nil:
			currentEndpoint.Description = extractValue(line)
		}
	}

	flushEndpoint()
	if currentGroup != nil {
		groups = append(groups, *currentGroup)
	}

	return groups, scanner.Err()
}

// extractValue pulls the value after the first colon on a markdown list line,
// stripping surrounding backticks and whitespace.
func extractValue(line string) string {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return ""
	}
	val := strings.TrimSpace(line[idx+1:])
	val = strings.Trim(val, "`")
	return strings.TrimSpace(val)
}

func DocumentationPageHandler(c *gin.Context) {
	groups, err := parseEndpointGroups("README.md")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read documentation")
		return
	}

	c.HTML(http.StatusOK, "documentation.html", gin.H{
		"Groups": groups,
	})
}
