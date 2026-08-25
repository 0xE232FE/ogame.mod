package wrapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnwrapTraderJSON(t *testing.T) {
	raw := []byte(`{"content":{"trader":"<div id=\"div_auctioneer\">ok</div>"},"newAjaxToken":"abc"}`)
	html, token := parseTraderAjax(raw)
	assert.Equal(t, `<div id="div_auctioneer">ok</div>`, string(html))
	assert.Equal(t, "abc", token)
	assert.Equal(t, `<div id="div_auctioneer">ok</div>`, string(unwrapTraderJSON(raw)))

	plain := []byte("<html>plain</html>")
	assert.Equal(t, plain, unwrapTraderJSON(plain))
}

func TestExtractTraderAjaxToken(t *testing.T) {
	page := []byte("    token = \"71136833af3e77439b3855df998672f3\"\n    initTrader();")
	assert.Equal(t, "71136833af3e77439b3855df998672f3", extractTraderAjaxToken(page))
}
