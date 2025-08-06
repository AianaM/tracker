package tracker

import (
	"net/http"
	"time"

	"github.com/AianaM/timefns"
)

// записи о затраченном времени
type Worklog struct {
	Self      string `json:"self"`
	ID        int    `json:"id"`
	Version   int    `json:"version"`
	Issue     Issue  `json:"issue"`
	Comment   string `json:"comment"`
	CreatedBy User   `json:"createdBy"`
	UpdatedBy User   `json:"updatedBy"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Start     string `json:"start"`
	Duration  string `json:"duration"`
}

type Issue struct {
	Self    string `json:"self"`
	Id      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display"`
}

type User struct {
	Self    string `json:"self"`
	Id      string `json:"id"`
	Display string `json:"display"`
}

func (t *TrackerClient) GetWorklog(createdBy string, createdAt timefns.TimeSpan) ([]Worklog, error) {
	return requestData[[]Worklog]{
		client: t,
		request: request{
			path:   "worklog/",
			method: http.MethodGet,
			params: []keyValue{
				{"createdBy", createdBy},
				{"createdAt", "from:" + createdAt.Start.Format(time.RFC3339Nano)},
				{"createdAt", "to:" + createdAt.End.Format(time.RFC3339Nano)},
			},
		},
	}.requestNew()
}
