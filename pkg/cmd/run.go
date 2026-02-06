package cmd

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

// Embedded static files
//
//go:embed ui/*
var content embed.FS

// SPAHandler handles Static files and SPA routing
type SPAHandler struct {
	staticFS fs.FS
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	if !strings.HasPrefix(requestPath, "/") {
		requestPath = "/" + requestPath
	}
	requestPath = path.Clean(requestPath)

	// 1. Try serving exact path
	if h.tryServe(w, r, requestPath) {
		return
	}

	// 2. Try serving with .html extension (for Next.js export routes)
	if h.tryServe(w, r, requestPath+".html") {
		return
	}

	// 3. Try serving path/index.html (for directories)
	if h.tryServe(w, r, path.Join(requestPath, "index.html")) {
		return
	}

	// 4. Fallback: 404
	// Try serving 404.html if it exists, with 404 status
	// We use tryServeWithStatus to ensure 404 code is sent
	if h.tryServeWithStatus(w, r, "/404.html", http.StatusNotFound) {
		return
	}

	http.NotFound(w, r)
}

func (h *SPAHandler) tryServe(w http.ResponseWriter, r *http.Request, p string) bool {
	return h.tryServeWithStatus(w, r, p, 0)
}

func (h *SPAHandler) tryServeWithStatus(w http.ResponseWriter, r *http.Request, p string, status int) bool {
	f, err := h.staticFS.Open(strings.TrimPrefix(p, "/"))
	if err != nil {
		return false
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil || stat.IsDir() {
		return false
	}

	if status != 0 {
		w.WriteHeader(status)
		// If we set a status code (like 404), we can't use ServeContent effectively
		// because it might try to set status 200 or handle Range requests which conflicts.
		// Instead, we just copy the content.
		_, err = io.Copy(w, f)
		return err == nil
	}

	http.ServeContent(w, r, p, stat.ModTime(), f.(io.ReadSeeker))
	return true
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the Sapliy Automation Studio locally",
	Long:  `Hosts the self-contained Sapliy Automation Studio web interface locally and proxies API requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		apiURL, _ := cmd.Flags().GetString("api")

		fmt.Printf("🚀 Sapliy Automation Studio starting...\n")
		fmt.Printf("   ├── UI: http://localhost:%s\n", port)
		fmt.Printf("   └── API Proxy: %s\n", apiURL)

		// Prepare FS
		fsys, err := fs.Sub(content, "ui")
		if err != nil {
			log.Fatal(err)
		}

		// API Proxy Handler
		target, err := url.Parse(apiURL)
		if err != nil {
			log.Fatal(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(target)

		// Mux
		mux := http.NewServeMux()

		// Handle API
		mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
			r.Host = target.Host
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
			proxy.ServeHTTP(w, r)
		})

		// Handle UI
		mux.Handle("/", &SPAHandler{staticFS: fsys})

		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("port", "p", "3000", "Port to serve the studio on")
	runCmd.Flags().StringP("api", "a", "http://localhost:8080", "Backend API URL to proxy to")
}
