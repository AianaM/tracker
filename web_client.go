package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/AianaM/timefns"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed templates/assets/*
var assetsFS embed.FS

const (
	serverPort     = ":8080"
	defaultWorklog = worklogPathPrefix + "/today"

	worklogPathPrefix = "/worklog"

	// Template files
	indexTemplate   = "templates/index.html"
	worklogTemplate = "templates/worklog.html"
)

type Page[T any] struct {
	Title   string
	Content T
}

type PageWorklogContent struct {
	Title    string
	Timespan timefns.TimeSpan
	Worklogs TableData
}

type PageWorklog Page[PageWorklogContent]

type TimeSpanTitled struct {
	title    string
	timespan timefns.TimeSpan
}

var handlers = map[string]http.HandlerFunc{
	"GET /worklog/{worklogPreset}":                                                    worklogHandler,
	"GET /worklog/{worklogPreset}/show/{showPreset}":                                  worklogHandler,
	"GET /worklog/{worklogPreset}/show/from/{showFrom}/to/{showTo}":                   worklogHandler,
	"GET /worklog/from/{worklogFrom}/to/{worklogTo}":                                  worklogHandler,
	"GET /worklog/from/{worklogFrom}/to/{worklogTo}/show/{showPreset}":                worklogHandler,
	"GET /worklog/from/{worklogFrom}/to/{worklogTo}/show/from/{showFrom}/to/{showTo}": worklogHandler,
}

func newWebClient() {
	// Check if cloud client is initialized
	if c.orgId == "" {
		log.Fatal("Cloud client not initialized. Make sure makeClouds() is called in init()")
	}

	log.Println("Starting server on", serverPort, "https://localhost"+serverPort)

	// Обслуживаем статические файлы из встроенной файловой системы
	assetsSubFS, err := fs.Sub(assetsFS, "templates/assets")
	if err != nil {
		log.Fatal("Error creating assets sub filesystem:", err)
	}
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsSubFS))))

	http.HandleFunc("/", viewHandler)
	for pattern, handler := range handlers {
		http.HandleFunc(pattern, handler)
	}
	log.Fatal(http.ListenAndServe(serverPort, nil))
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// Redirect to the worklog page
		http.Redirect(w, r, defaultWorklog, http.StatusFound)
		return
	}
	http.NotFound(w, r)
	log.Printf("Invalid path: %s", r.URL.Path)
}

func worklogHandler(w http.ResponseWriter, r *http.Request) {
	var worklogTimespan TimeSpanTitled
	var showTimespan TimeSpanTitled

	if preset := r.PathValue("worklogPreset"); preset != "" {
		worklog, err := parsePresetTimespan(preset)
		if err != nil {
			log.Printf("Error handling preset worklogPreset: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		worklogTimespan = worklog
	} else if r.PathValue("worklogFrom") != "" && r.PathValue("worklogTo") != "" {
		worklog, err := parseTimeSpan(r.PathValue("worklogFrom"), r.PathValue("worklogTo"))
		if err != nil {
			log.Printf("error parsing worklogTimespan: %v", err)
			http.Error(w, "Invalid worklogTimespan", http.StatusBadRequest)
			return
		}
		worklogTimespan = TimeSpanTitled{"Custom", worklog}
	} else {
		http.NotFound(w, r)
		log.Println("Invalid worklog path:", r.URL.Path)
		return
	}

	if preset := r.PathValue("showPreset"); preset != "" {
		show, err := parsePresetTimespan(preset)
		if err != nil {
			log.Printf("Error handling preset showPreset: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		showTimespan = show
	} else if r.PathValue("showFrom") != "" && r.PathValue("showTo") != "" {
		show, err := parseTimeSpan(r.PathValue("showFrom"), r.PathValue("showTo"))
		if err != nil {
			log.Printf("error parsing showTimespan: %w", err)
			http.Error(w, "Invalid showTimespan", http.StatusBadRequest)
			return
		}
		showTimespan = TimeSpanTitled{"Custom", show}
	} else {
		// Default to showing the same timespan as worklog
		showTimespan = worklogTimespan
	}

	page, err := createWorklogPage(worklogTimespan.timespan, showTimespan)
	if err != nil {
		log.Printf("Error creating worklog page: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	page.render(w, r)
}

func parsePresetTimespan(preset string) (TimeSpanTitled, error) {
	var title string
	var timespan timefns.TimeSpan

	switch preset {
	case "today":
		title = "Today"
		timespan = timefns.Today()
	case "currentWeek":
		title = "Current Week"
		timespan = timefns.CurrentWeek()
	case "currentMonth":
		title = "Current Month"
		timespan = timefns.CurrentMonth()
	default:
		return TimeSpanTitled{}, fmt.Errorf("unknown preset: %s", preset)
	}

	return TimeSpanTitled{title, timespan}, nil
}

func parseTimeSpan(start, end string) (timefns.TimeSpan, error) {
	startDate, err := time.Parse(time.DateOnly, start)
	if err != nil {
		return timefns.TimeSpan{}, fmt.Errorf("error parsing start date: %w", err)
	}
	endDate, err := time.Parse(time.DateOnly, end)
	if err != nil {
		return timefns.TimeSpan{}, fmt.Errorf("error parsing end date: %w", err)
	}
	return timefns.TimeSpan{Start: startDate, End: endDate}, nil
}

func createWorklogPage(timespan timefns.TimeSpan, show TimeSpanTitled) (PageWorklog, error) {
	worklogs, err := c.getWorklog(timespan)
	worklogsTable := worklogs.getWorklogsTable(show.timespan)
	if err != nil {
		return PageWorklog{}, fmt.Errorf("error getting worklogs: %w", err)
	}

	return PageWorklog{
		Title: show.title,
		Content: PageWorklogContent{
			Title:    show.title,
			Timespan: show.timespan,
			Worklogs: worklogsTable,
		},
	}, nil
}

func (page PageWorklog) render(w http.ResponseWriter, r *http.Request) {
	log.Printf("Rendering page: %s with timespan: %v - %v", page.Title, page.Content.Timespan.Start, page.Content.Timespan.End)
	t, err := template.New("index.html").Funcs(template.FuncMap{
		"durationBeautify": DurationBeautify,
		"trackerUrl":       TrackerUrl,
	}).ParseFS(templatesFS, indexTemplate, worklogTemplate)
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, page); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
