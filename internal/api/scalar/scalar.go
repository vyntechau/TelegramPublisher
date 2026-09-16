package scalar

import (
	_ "embed"
	"net/http"
	"os"
)

// ScalarHandler serves the modern Scalar API Documentation UI.
func Register(mux *http.ServeMux, openAPIPath string) {
	mux.HandleFunc("/docs/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := os.ReadFile(openAPIPath)
		if err != nil {
			data, _ = os.ReadFile("docs/openapi.json")
		}
		_, _ = w.Write(data)
	})

	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `<!doctype html>
<html>
  <head>
    <title>TelegramPublisher API Documentation (Scalar)</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link rel="icon" type="image/svg+xml" href="https://scalar.com/favicon.svg" />
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/docs/openapi.json"
      data-configuration='{"theme": "deepSpace", "layout": "modern"}'></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`
		_, _ = w.Write([]byte(html))
	})
}
