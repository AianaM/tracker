package worklog

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"example.com/tracker/internal/tracker"
	"github.com/AianaM/timefns"
)

const (
	name       = "worklog"
	pathPrefix = "/" + name
	isDev      = false
)

//go:embed static/*
var StaticFiles embed.FS

//go:embed templates/*
var TemplatesFs embed.FS

type templateConfig struct {
	name    string
	funcMap template.FuncMap
	css     template.CSS
	tpl     *template.Template
}
type Handler struct {
	trackerClient *tracker.TrackerClient
	templates     templateConfig
}

type timespanParams struct {
	Preset string
	From   string
	To     string
}

var pathParams = struct {
	Worklog timespanParams
	Show    timespanParams
}{
	Worklog: timespanParams{
		Preset: "worklogPreset",
		From:   "worklogFrom",
		To:     "worklogTo",
	},
	Show: timespanParams{
		Preset: "showPreset",
		From:   "showFrom",
		To:     "showTo",
	},
}

type titledTimeSpan struct {
	title    string
	timespan timefns.TimeSpan
}

type PageWorklogContent struct {
	Title    string
	Timespan timefns.TimeSpan
	Worklogs TableData
	Style    template.CSS
}

type PageWorklog struct {
	Title   string
	Content PageWorklogContent
}

func NewHandler(trackerClient *tracker.TrackerClient, indexTpl *template.Template) (*Handler, error) {
	funcMap := getFuncMap(trackerClient.HostURL)
	css, err := getStyle()
	if err != nil {
		return nil, fmt.Errorf("error getting style: %w", err)
	}
	tpl, err := getTpl(funcMap, indexTpl)
	if err != nil {
		return nil, fmt.Errorf("error getting template: %w", err)
	}

	tpls := templateConfig{
		name:    tpl.Name(),
		funcMap: funcMap,
		css:     css,
		tpl:     tpl,
	}

	return &Handler{
		trackerClient: trackerClient,
		templates:     tpls,
	}, nil
}

func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+pathPrefix+"/{"+pathParams.Worklog.Preset+"}", h.worklogHandler)
	mux.HandleFunc("GET "+pathPrefix+"/{"+pathParams.Worklog.Preset+"}/show/{"+pathParams.Show.Preset+"}", h.worklogShowHandler)
	mux.HandleFunc("GET "+pathPrefix+"/{"+pathParams.Worklog.Preset+"}/show/from/{"+pathParams.Show.From+"}/to/{"+pathParams.Show.To+"}", h.worklogShowHandler)
	mux.HandleFunc("GET "+pathPrefix+"/from/{"+pathParams.Worklog.From+"}/to/{"+pathParams.Worklog.To+"}", h.worklogHandler)
	mux.HandleFunc("GET "+pathPrefix+"/from/{"+pathParams.Worklog.From+"}/to/{"+pathParams.Worklog.To+"}/show/{"+pathParams.Show.Preset+"}", h.worklogShowHandler)
	mux.HandleFunc("GET "+pathPrefix+"/from/{"+pathParams.Worklog.From+"}/to/{"+pathParams.Worklog.To+"}/show/from/{"+pathParams.Show.From+"}/to/{"+pathParams.Show.To+"}", h.worklogShowHandler)
}

func (h *Handler) HandleStatic(mux *http.ServeMux) {
	assetsSubFS, err := fs.Sub(StaticFiles, "static")
	if err != nil {
		log.Fatal("Error creating assets sub filesystem:", err)
	}
	mux.Handle("/"+name+"/", http.StripPrefix("/"+name+"/", http.FileServer(http.FS(assetsSubFS))))
}

func (h *Handler) worklogHandler(w http.ResponseWriter, r *http.Request) {
	worklog, err := pathParams.Worklog.parse(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing worklog parameters: %v", err), http.StatusBadRequest)
		return
	}
	h.render(worklog, worklog, w)
}

func (h *Handler) worklogShowHandler(w http.ResponseWriter, r *http.Request) {
	worklog, err := pathParams.Worklog.parse(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing worklog parameters: %v", err), http.StatusBadRequest)
		return
	}
	show, err := pathParams.Show.parse(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing show parameters: %v", err), http.StatusBadRequest)
		return
	}
	h.render(worklog, show, w)
}

func parsePresetTimespan(preset string) (titledTimeSpan, error) {
	switch preset {
	case "today":
		return titledTimeSpan{"Сегодня", timefns.Today()}, nil
	case "currentWeek":
		return titledTimeSpan{"Текущая неделя", timefns.CurrentWeek()}, nil
	case "currentMonth":
		return titledTimeSpan{"Текущий месяц", timefns.CurrentMonth()}, nil
	default:
		return titledTimeSpan{}, fmt.Errorf("unknown preset: %s", preset)
	}
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

func (t timespanParams) parse(r *http.Request) (titledTimeSpan, error) {
	if preset := r.PathValue(t.Preset); preset != "" {
		worklog, err := parsePresetTimespan(preset)
		if err != nil {
			return titledTimeSpan{}, fmt.Errorf("error handling preset %v: %v", t.Preset, err)
		}
		return worklog, nil
	} else if from, to := r.PathValue(t.From), r.PathValue(t.To); from != "" && to != "" {
		worklog, err := parseTimeSpan(r.PathValue(t.From), r.PathValue(t.To))
		if err != nil {
			return titledTimeSpan{}, fmt.Errorf("error handling custom %v, %v: %v", t.From, t.To, err)
		}
		return titledTimeSpan{"Кастомный", worklog}, nil
	} else {
		return titledTimeSpan{}, fmt.Errorf("invalid worklog path: %v", r.URL.Path)
	}
}

func (h *Handler) render(worklog, show titledTimeSpan, w http.ResponseWriter) {
	page, err := h.createWorklogPage(worklog, show)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating worklog page: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.tpl.ExecuteTemplate(w, "index.html", page); err != nil {
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), 500)
	}
}

func (h *Handler) createWorklogPage(timespan, show titledTimeSpan) (PageWorklog, error) {
	createdBy := os.Getenv("LOGIN") // TODO: params
	worklogsTable, err := h.getWorklogsTable(createdBy, timespan, show)
	if err != nil {
		return PageWorklog{}, fmt.Errorf("error getting worklogs: %w", err)
	}

	return PageWorklog{
		Title: show.title,
		Content: PageWorklogContent{
			Title:    show.title,
			Timespan: show.timespan,
			Worklogs: worklogsTable,
			Style:    h.templates.css,
		},
	}, nil
}

func DurationBeautify(d time.Duration) string {
	h := int(d / time.Hour)
	m := int((d % time.Hour) / time.Minute)

	parts := make([]string, 0, 2)
	if h > 0 {
		parts = append(parts, strconv.Itoa(h)+"h")
	}

	if m > 0 {
		parts = append(parts, strconv.Itoa(m)+"m")
	}

	if len(parts) == 0 {
		return "0m"
	}

	return strings.Join(parts, " ")
}
func getFuncMap(hostURL string) template.FuncMap {
	return map[string]interface{}{
		"durationBeautify": DurationBeautify,
		"trackerUrl": func(issueKey string) string {
			if hostURL == "" {
				return ""
			}
			if strings.HasSuffix(hostURL, "/") {
				return hostURL + issueKey
			}
			return hostURL + "/" + issueKey
		},
	}
}
func getStyle() (template.CSS, error) {
	var style []byte
	var err error
	if isDev {
		style, err = os.ReadFile("internal/worklog/static/css/style.css")
	} else {
		style, err = fs.ReadFile(StaticFiles, "static/css/style.css")
	}
	if err != nil {
		return template.CSS(""), fmt.Errorf("error reading style file: %v", err)
	}
	return template.CSS(style), nil
}
func getTpl(funcMap template.FuncMap, indexTpl *template.Template) (*template.Template, error) {
	var w *template.Template
	if isDev {
		w = template.Must(template.Must(indexTpl.Clone()).New("worklog.html").Funcs(funcMap).ParseFiles("internal/worklog/templates/worklog.html"))
	} else {
		w = template.Must(template.Must(indexTpl.Clone()).New("worklog.html").Funcs(funcMap).ParseFS(TemplatesFs, "templates/worklog.html"))
	}

	return w.Lookup("index.html"), nil
}
