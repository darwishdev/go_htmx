package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"sync"
)

//go:embed static
var staticFS embed.FS

//go:embed templates
var tmplFS embed.FS

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

	// Each route is parsed together with the shared layout + nav/footer
	// partials, so executing "layout" yields the full page and "content"
	// yields just that route's body (the HTMX-swapped fragment).
	homeTmpl  = parsePage("home.html")
	aboutTmpl = parsePage("about.html")
	// countTmpl renders just the counter fragment that HTMX swaps in.
	countTmpl = template.Must(template.ParseFS(tmplFS, "templates/count.html"))
)

// parsePage builds a template set from the layout, the shared partials, the
// count fragment, and one route file — the Go analog of a Vue route reusing
// the app layout and shared components.
func parsePage(name string) *template.Template {
	return template.Must(template.ParseFS(tmplFS,
		"templates/layout.html",
		"templates/partials.html",
		"templates/count.html",
		"templates/"+name,
	))
}

// pageData is the view model handed to every route template.
type pageData struct {
	Title  string
	Active string // which nav link to highlight: "home" | "about"
	Count  int
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", handleHome)
	mux.HandleFunc("GET /about", handleAbout)
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

func handleHome(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, homeTmpl, pageData{Title: "Home", Active: "home", Count: count.value()})
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, aboutTmpl, pageData{Title: "About", Active: "about"})
}

func handleDelta(delta int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderFragment(w, countTmpl, "count", pageData{Count: count.add(delta)})
	}
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	renderFragment(w, countTmpl, "count", pageData{Count: count.add(-count.value())})
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

// renderPage writes a full document for a normal browser request, but only the
// "content" block when HTMX asks for it (HX-Request header) — so nav clicks
// swap just the middle of the page instead of reloading the whole shell.
func renderPage(w http.ResponseWriter, r *http.Request, t *template.Template, data pageData) {
	name := "layout"
	if r.Header.Get("HX-Request") == "true" {
		name = "content"
	}
	renderFragment(w, t, name, data)
}

func renderFragment(w http.ResponseWriter, t *template.Template, name string, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
