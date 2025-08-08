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
type Preset string
type timespanParams struct {
	Preset string
	From   string
	To     string
}
type PathParams struct {
	Worklog   timespanParams
	Show      timespanParams
	CreatedBy string
}

var pathParams = PathParams{
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
	CreatedBy: "createdBy",
}

type titledTimeSpan[T any] struct {
	Title    string
	Timespan struct {
		Start T
		End   T
	}
}
type Query[T any] struct {
	CreatedBy       string
	CreatedAt, Show titledTimeSpan[T]
}
type PageWorklogContent struct {
	Query    Query[string]
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
	mux.HandleFunc("GET "+pathPrefix+"/{"+pathParams.CreatedBy+"}/{"+pathParams.Worklog.Preset+"}", h.worklogHandler44(worklogQuery))
	mux.HandleFunc("GET "+pathPrefix+"/{"+pathParams.CreatedBy+"}/{"+pathParams.Worklog.Preset+"}/show/{"+pathParams.Show.Preset+"}", h.worklogHandler44(worklogShowQuery))
	mux.HandleFunc("GET "+pathPrefix+"/{"+pathParams.CreatedBy+"}/{"+pathParams.Worklog.Preset+"}/show/from/{"+pathParams.Show.From+"}/to/{"+pathParams.Show.To+"}", h.worklogHandler44(worklogShowQuery))
	mux.HandleFunc("GET "+pathPrefix+"/{"+pathParams.CreatedBy+"}/from/{"+pathParams.Worklog.From+"}/to/{"+pathParams.Worklog.To+"}", h.worklogHandler44(worklogQuery))
	mux.HandleFunc("GET "+pathPrefix+"/{"+pathParams.CreatedBy+"}/from/{"+pathParams.Worklog.From+"}/to/{"+pathParams.Worklog.To+"}/show/{"+pathParams.Show.Preset+"}", h.worklogHandler44(worklogShowQuery))
	mux.HandleFunc("GET "+pathPrefix+"/{"+pathParams.CreatedBy+"}/from/{"+pathParams.Worklog.From+"}/to/{"+pathParams.Worklog.To+"}/show/from/{"+pathParams.Show.From+"}/to/{"+pathParams.Show.To+"}", h.worklogHandler44(worklogShowQuery))
}

func (h *Handler) HandleStatic(mux *http.ServeMux) {
	assetsSubFS, err := fs.Sub(StaticFiles, "static")
	if err != nil {
		log.Fatal("Error creating assets sub filesystem:", err)
	}
	mux.Handle("GET /"+name+"/js/{fileName}", http.StripPrefix("/"+name+"/", http.FileServer(http.FS(assetsSubFS))))
}

func activatedRoute(r *http.Request) *PathParams {
	return &PathParams{
		CreatedBy: r.PathValue(pathParams.CreatedBy),
		Worklog: timespanParams{
			Preset: r.PathValue(pathParams.Worklog.Preset),
			From:   r.PathValue(pathParams.Worklog.From),
			To:     r.PathValue(pathParams.Worklog.To),
		},
		Show: timespanParams{
			Preset: r.PathValue(pathParams.Show.Preset),
			From:   r.PathValue(pathParams.Show.From),
			To:     r.PathValue(pathParams.Show.To),
		},
	}
}
func worklogQuery(p PathParams) (*Query[time.Time], error) {
	if p.CreatedBy == "" {
		return nil, fmt.Errorf("Error parsing CreatedBy parameter")
	}
	createdAt, err := p.Worklog.parse()
	if err != nil {
		return nil, fmt.Errorf("Error parsing worklog parameters: %v", err)
	}
	return &Query[time.Time]{
		CreatedBy: p.CreatedBy,
		CreatedAt: createdAt,
		Show:      createdAt,
	}, nil
}

func worklogShowQuery(p PathParams) (*Query[time.Time], error) {
	q, err := worklogQuery(p)
	if err != nil {
		return nil, fmt.Errorf("Error creating worklog query: %v", err)
	}
	show, err := p.Show.parse()
	if err != nil {
		return nil, fmt.Errorf("Error parsing show parameters: %v", err)
	}
	q.Show = show
	return q, nil
}

func (h *Handler) worklogHandler44(queryFn func(p PathParams) (*Query[time.Time], error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		activatedRoute := activatedRoute(r)
		q, err := queryFn(*activatedRoute)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating worklog query: %v", err), http.StatusInternalServerError)
			return
		}
		page, err := h.createWorklogPage(*q)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating worklog page: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := h.templates.tpl.ExecuteTemplate(w, "index.html", page); err != nil {
			http.Error(w, fmt.Sprintf("Template execution error: %v", err), 500)
		}
	}
}

func (p Preset) parsePresetTimespan() (titledTimeSpan[time.Time], error) {
	switch p {
	case "today":
		return titledTimeSpan[time.Time]{"Сегодня", timefns.Today()}, nil
	case "currentWeek":
		return titledTimeSpan[time.Time]{"Текущая неделя", timefns.CurrentWeek()}, nil
	case "currentMonth":
		return titledTimeSpan[time.Time]{"Текущий месяц", timefns.CurrentMonth()}, nil
	default:
		return titledTimeSpan[time.Time]{}, fmt.Errorf("unknown preset: %s", p)
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
func (t timespanParams) parse() (titledTimeSpan[time.Time], error) {
	if t.Preset != "" {
		if worklog, err := Preset(t.Preset).parsePresetTimespan(); err != nil {
			return titledTimeSpan[time.Time]{}, fmt.Errorf("error parsing preset %v: %v", t, err)
		} else {
			return worklog, nil
		}
	} else if t.From != "" && t.To != "" {
		if worklog, err := parseTimeSpan(t.From, t.To); err != nil {
			return titledTimeSpan[time.Time]{}, fmt.Errorf("error parsing custom %v: %v", t, err)
		} else {
			return titledTimeSpan[time.Time]{"Кастомный", worklog}, nil
		}
	}
	return titledTimeSpan[time.Time]{}, fmt.Errorf("error parsing worklog path: %v", t)
}

func (h *Handler) createWorklogPage(q Query[time.Time]) (PageWorklog, error) {
	worklogsTable, err := h.getWorklogsTable(q.CreatedBy, q.CreatedAt, q.Show)
	if err != nil {
		return PageWorklog{}, fmt.Errorf("error getting worklogs: %w", err)
	}
	log.Println(q.CreatedAt.Timespan.Start.Format(timefns.ISO8601n))

	return PageWorklog{
		Title: "Worklog: " + q.Show.Title,
		Content: PageWorklogContent{
			Query: Query[string]{
				CreatedBy: q.CreatedBy,
				CreatedAt: titledTimeSpan[string]{
					Title: q.CreatedAt.Title,
					Timespan: struct {
						Start string
						End   string
					}{Start: q.CreatedAt.Timespan.Start.Format(time.DateOnly), End: q.CreatedAt.Timespan.End.Format(time.DateOnly)}},
				Show: titledTimeSpan[string]{
					Title: q.Show.Title,
					Timespan: struct {
						Start string
						End   string
					}{Start: q.Show.Timespan.Start.Format(time.DateOnly), End: q.Show.Timespan.End.Format(time.DateOnly)}},
			},
			Worklogs: worklogsTable,
			Style:    h.templates.css,
		}}, nil
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
