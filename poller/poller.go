package poller

import (
	"context"
	"net/http"

	"google.golang.org/appengine/log"
)

// weekDay는 요일을 나타냅니다.
type weekDay int

const (
	Sunday    weekDay = iota // 일요일, 0
	Monday                   // 월요일, 1
	Tuesday                  // 화요일, 2
	Wednesday                // 수요일, 3
	Thursday                 // 목요일, 4
	Friday                   // 금요일, 5
	Saturday                 // 토요일, 6
	Others                   // 기타, 7
	New                      // 신작, 8
)

// AnimeInfo 는 애니메이션 정보를 나타냅니다.
type AnimeInfo struct {
	animeNo      int    `json:"animeNo"`
	week         int    `json:"week"`
	status       string `json:"status"`
	time         string `json:"time"`
	subject      string `json:"subject"`
	genres       string `json:"genres"`
	captionCount int    `json:"captionCount"`
	startDate    string `json:"startDate"`
	endDate      string `json:"endDate"`
	website      string `json:"website"`
}

// Poller는 주기적으로 API를 통해서 신작 애니메이션 편성표 정보를 받아옵니다.
// 신작 애니메이션 No로 자막 정보를 수집합니다.
// 애니메이션 자막이 업로드되면 scraper에 자막 수집 요청을 보냅니다.
type Poller struct {
	pollingInterval int // Polling interval in minutes
}

// NewPoller는 Poller를 생성합니다.
func NewPoller(pollingInterval int) *Poller {
	return &Poller{
		pollingInterval,
	}
}

// Run은 Poller를 실행합니다.
func (p *Poller) Run() {

}

// 신작 애니메이션 편성표 정보를 받아옵니다.
func (p *Poller) getNewAnimeSchedule(ctx context.Context) {
	res, err := http.Get("https://api.manatoki95.net/webtoon/weekday")
	if err != nil {
		log.Errorf(ctx, "Failed to get new anime schedule: %v", err)
	}
	defer res.Body.Close()
}
