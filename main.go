package main

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"

	"example.com/tracker/internal/client"
	"example.com/tracker/internal/config"
	"example.com/tracker/internal/server"
	"example.com/tracker/internal/tracker"
	"example.com/tracker/internal/worklog"
	"example.com/tracker/web"
)

func init() {
	if loc, err := time.LoadLocation("Europe/Moscow"); err != nil {
		fmt.Println(err)
	} else {
		time.Local = loc
	}
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create HTTP client with interceptors
	httpClient := client.New([]client.Interceptor{
		tracker.AuthTokenInterceptor(cfg.YandexIAMToken, cfg.YandexOrgID),
		// client.LoggingInterceptor(), // uncomment for debugging
	})

	// Create tracker client
	trackerClient := tracker.NewTrackerClient(tracker.Config{
		HostURL: cfg.TrackerHost,
		Client:  httpClient,
		Ctx:     context.Background(),
		Timeout: 10 * time.Second,
	})

	// Load templates
	indexTpl := template.Must(template.ParseFS(web.Templates, "templates/index.html"))

	// Create worklog handler
	worklogHandler, err := worklog.NewHandler(trackerClient, indexTpl)
	if err != nil {
		log.Fatalf("Failed to create worklog handler: %v", err)
	}

	// Setup routes
	mux := http.NewServeMux()

	// Root redirect
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/worklog/today", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	// Worklog routes
	worklogHandler.SetupRoutes(mux)
	worklogHandler.HandleStatic(mux)

	// Static files
	handleStatic(mux)

	// Apply middleware
	handler := server.WithLogging(
		server.WithCORS(mux),
	)

	// Start server
	log.Printf("Starting server on %s", cfg.ServerAddr)
	server.StartServer(handler, cfg.ServerAddr)
}

func handleStatic(mux *http.ServeMux) {
	assetsSubFS, err := fs.Sub(web.StaticFiles, "static")
	if err != nil {
		log.Fatal("Error creating assets sub filesystem:", err)
	}
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsSubFS))))
}
