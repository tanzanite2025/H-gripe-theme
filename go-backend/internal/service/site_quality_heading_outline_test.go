package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestWalkSiteQualityHeadingHTMLSkipsHiddenAndClosedDetails(t *testing.T) {
	root, err := html.Parse(strings.NewReader(`<!doctype html>
<html>
  <body>
    <h1>Visible hero</h1>
    <div hidden><h2>Hidden hidden</h2></div>
    <div inert><h2>Hidden inert</h2></div>
    <div aria-hidden="true"><h2>Hidden aria</h2></div>
    <div style="visibility: collapse"><h2>Hidden collapse</h2></div>
    <div style="content-visibility: hidden"><h2>Hidden content visibility</h2></div>
    <template><h2>Hidden template</h2></template>
    <details>
      <h2>Hidden details body</h2>
      <summary><h2>Visible summary</h2></summary>
    </details>
    <details open>
      <h2>Visible open details</h2>
    </details>
  </body>
</html>`))
	require.NoError(t, err)

	headings := make([]siteQualityHeadingNode, 0)
	walkSiteQualityHeadingHTML(root, nil, &headings, false)

	texts := make([]string, 0, len(headings))
	for _, heading := range headings {
		texts = append(texts, heading.Text)
	}

	require.Equal(t, []string{
		"Visible hero",
		"Visible summary",
		"Visible open details",
	}, texts)
}
