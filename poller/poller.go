package poller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

// AnimeInfo 구조체는 각 애니메이션에 대한 정보를 정의합니다.
type AnimeInfo struct {
	Week         string `json:"week"`
	AnimeNo      int    `json:"animeNo"`
	Status       string `json:"status"`
	Time         string `json:"time"`
	Subject      string `json:"subject"`
	Genres       string `json:"genres"`
	CaptionCount int    `json:"captionCount"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	Website      string `json:"website"`
}

// AnimeSchedule 구조체는 전체 응답을 정의합니다.
type AnimeSchedule struct {
	Code string      `json:"code"`
	Data []AnimeInfo `json:"data"`
}

// Poller는 주기적으로 API를 통해서 신작 애니메이션 편성표 정보를 받아옵니다.
// 신작 애니메이션 No로 자막 정보를 수집합니다.
// 애니메이션 자막이 업로드되면 scraper에 자막 수집 요청을 보냅니다.
type Poller struct {
	pollingInterval string // Polling interval
}

// NewPoller는 Poller를 생성합니다.
func NewPoller(pollingInterval string) *Poller {
	return &Poller{
		pollingInterval,
	}
}

// Run은 Poller를 실행합니다.
func (p *Poller) Run() {
	// 1. 신작 애니메이션 편성표 정보를 받아옵니다.
	animeInfos, err := p.getNewAnimeSchedule()
	if err != nil {
		log.Printf("failed to get new anime schedule: %v", err)
		return
	}
	log.Printf("AnimeCount: %d", len(animeInfos))
	// 2. 신작 애니메이션 편성표 정보를 DB에 저장합니다.
	// 3. 신작 애니메이션 No로 자막 정보를 수집합니다.
	// 4. 애니메이션 자막이 업로드되면 scraper에 자막 수집 요청을 보냅니다.
}

// 신작 애니메이션 편성표 정보를 받아옵니다.
func (p *Poller) getNewAnimeSchedule() ([]AnimeInfo, error) {
	var animeInfos []AnimeInfo
	// 요일별로 신작 애니메이션 편성표 정보를 받아옵니다.
	for _, day := range []weekDay{Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Others} {
		reqUrl := fmt.Sprintf("https://api.anissia.net/anime/schedule/%d", day)
		res, err := http.Get(reqUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to get new anime schedule: %v", err)
		}
		defer res.Body.Close()

		// Check response status code
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get new anime schedule: %v", err)
		}

		// JSON을 파싱합니다.
		var schedule AnimeSchedule
		err = json.NewDecoder(res.Body).Decode(&schedule)
		if err != nil {
			return nil, fmt.Errorf("failed to decode response body: %v", err)
		}
		for _, animeInfo := range schedule.Data {
			if (animeInfo.CaptionCount > 0) && (animeInfo.Status == "ON") {
				animeInfos = append(animeInfos, animeInfo)
			}
		}
	}
	return animeInfos, nil
}
