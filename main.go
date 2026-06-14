package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
)

//go:embed static
var staticFS embed.FS

// counter is a tiny in-memory, concurrency-safe reactive value.
type counter struct {
	mu sync.Mutex
	n  int
}

func (c *counter) add(delta int) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n += delta
	return c.n
}

func (c *counter) value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

var (
	count = &counter{}

	pageTmpl = template.Must(template.New("page").Parse(pageHTML))
	// countTmpl renders just the counter fragment that HTMX swaps in.
	countTmpl = template.Must(template.New("count").Parse(countHTML))
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	// Serve embedded static assets (icons) and PWA files at stable paths.
	mux.Handle("GET /static/", http.FileServerFS(staticFS))
	mux.HandleFunc("GET /manifest.json", serveStatic("static/manifest.json", "application/manifest+json"))
	mux.HandleFunc("GET /sw.js", serveStatic("static/sw.js", "text/javascript"))
	mux.HandleFunc("POST /increment", handleDelta(1))
	mux.HandleFunc("POST /decrement", handleDelta(-1))
	mux.HandleFunc("POST /reset", handleReset)

	addr := ":8080"
	log.Printf("listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	render(w, pageTmpl, count.value())
}

func handleDelta(delta int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, countTmpl, count.add(delta))
	}
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	render(w, countTmpl, count.add(-count.value()))
}

func serveStatic(path, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := staticFS.ReadFile(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(b)
	}
}

func render(w http.ResponseWriter, t *template.Template, n int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, strconv.Itoa(n)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const countHTML = `<span id="count">{{.}}</span>`

const pageHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Go + HTMX Counter</title>
  <meta name="theme-color" content="#0f172a">
  <link rel="manifest" href="/manifest.json">
  <link rel="icon" href="/static/icon-192.png">
  <link rel="apple-touch-icon" href="/static/icon-192.png">
  <script src="https://unpkg.com/htmx.org@2.0.4"></script>
  <script>
    if ('serviceWorker' in navigator) {
      window.addEventListener('load', () => navigator.serviceWorker.register('/sw.js'));
    }
  </script>
  <style>
    body { font-family: system-ui, sans-serif; display: grid; place-items: center;
           min-height: 100vh; margin: 0; background: #0f172a; color: #e2e8f0; }
    .card { text-align: center; padding: 2.5rem 3rem; background: #1e293b;
            border-radius: 16px; box-shadow: 0 10px 30px rgba(0,0,0,.4); }
    h1 { margin: 0 0 .25rem; font-size: 1.25rem; font-weight: 600; }
    p { margin: 0 0 1.5rem; color: #94a3b8; }
    .value { font-size: 4rem; font-weight: 700; font-variant-numeric: tabular-nums; }
    .buttons { display: flex; gap: .5rem; justify-content: center; margin-top: 1.5rem; }
    button { font-size: 1.1rem; padding: .6rem 1.1rem; border: none; border-radius: 10px;
             cursor: pointer; background: #334155; color: #e2e8f0; transition: background .15s; }
    button:hover { background: #475569; }
    button.primary { background: #6366f1; }
    button.primary:hover { background: #818cf8; }
  </style>
</head>
<body>
  <main class="card">
    <h1>Hello, world 👋</h1>
    <p>A reactive counter, powered by Go &amp; HTMX</p>
    <div class="value"><span id="count">{{.}}</span></div>
    <div class="buttons">
      <button hx-post="/decrement" hx-target="#count" hx-swap="outerHTML">−</button>
      <button hx-post="/reset" hx-target="#count" hx-swap="outerHTML">reset</button>
      <button class="primary" hx-post="/increment" hx-target="#count" hx-swap="outerHTML">+</button>
    </div>
  </main>
</body>
</html>`
