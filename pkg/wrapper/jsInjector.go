package wrapper

import (
	"os"
	"strings"
)

func injectJS(pageHTML []byte) []byte {
	stringHTML := string(pageHTML)

	jsByte, err := os.ReadFile("OGLight_v5.2.0-tbot.user.js")
	if err != nil {
		return []byte{}
	}

	stringHTML = strings.ReplaceAll(stringHTML, "</head>", "<script>"+string(jsByte)+"</script></head>")
	return []byte(stringHTML)
}
