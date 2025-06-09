package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/AianaM/timefns"
)

const (
	serverPort     = ":8080"
	templatesDir   = "templates/"
	assetsDir      = "templates/assets"
	defaultWorklog = worklogPathPrefix + "/today"

	worklogPathPrefix = "/worklog"

	// Template files
	indexTemplate   = "index.html"
	worklogTemplate = "worklog.html"
)

type Page[T any] struct {
	Title   string
	Content T
}

type WorklogPageContent struct {
	Title    string
	Timespan timefns.TimeSpan
	Worklogs TableData
}

var (
	presetPath  = regexp.MustCompile("^" + worklogPathPrefix + "/(today|currentWeek|currentMonth)$")
	worklogPath = regexp.MustCompile("^" + worklogPathPrefix + "/from/(?P<from>.+)/to/(?P<to>.+)$")
)

func newWebClient() {
	// Check if cloud client is initialized
	if c.orgId == "" {
		log.Fatal("Cloud client not initialized. Make sure makeClouds() is called in init()")
	}

	log.Println("Starting server on", serverPort, "https://localhost"+serverPort)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))
	http.HandleFunc("/", viewHandler)
	http.HandleFunc(worklogPathPrefix+"/", worklogHandler)
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
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var page Page[WorklogPageContent]

	worklogPresetMatches := presetPath.FindStringSubmatch(r.URL.Path)
	if worklogPresetMatches != nil {
		if v, err := handlePresetWorklog(worklogPresetMatches[1]); err != nil {
			log.Printf("Error handling preset worklog: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		} else {
			page = v
		}
	} else {
		worklogMatches := worklogPath.FindStringSubmatch(r.URL.Path)
		fromIndex := worklogPath.SubexpIndex("from")
		toIndex := worklogPath.SubexpIndex("to")
		if fromIndex > -1 && fromIndex < len(worklogMatches) && toIndex > -1 && toIndex < len(worklogMatches) {
			if from, err := parseTimeSpan(worklogMatches[fromIndex], worklogMatches[toIndex]); err != nil {
				log.Printf("Error parsing time span: %v", err)
				http.Error(w, "Invalid date range", http.StatusBadRequest)
				return
			} else {
				page, err = createWorklogPage(from, "Custom Worklog")
				if err != nil {
					log.Printf("Error creating worklog page: %v", err)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			}
		} else {
			http.NotFound(w, r)
			log.Println("Invalid worklog path:", r.URL.Path)
			return
		}
	}

	log.Printf("Rendering page: %s with timespan: %v - %v", page.Title, page.Content.Timespan.Start, page.Content.Timespan.End)
	t, err := template.New(indexTemplate).Funcs(template.FuncMap{
		"durationBeautify": DurationBeautify,
	}).ParseFiles(templatesDir+indexTemplate, templatesDir+worklogTemplate)
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

func handlePresetWorklog(preset string) (Page[WorklogPageContent], error) {
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
		return Page[WorklogPageContent]{}, fmt.Errorf("unknown preset: %s", preset)
	}

	return createWorklogPage(timespan, title)
}

func createWorklogPage(timespan timefns.TimeSpan, title string) (Page[WorklogPageContent], error) {
	worklogs, err := c.getWorklog(timespan)
	if err != nil {
		return Page[WorklogPageContent]{}, fmt.Errorf("error getting worklogs: %w", err)
	}

	return Page[WorklogPageContent]{
		Title: title,
		Content: WorklogPageContent{
			Title:    title,
			Timespan: timespan,
			Worklogs: worklogs,
		},
	}, nil
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
