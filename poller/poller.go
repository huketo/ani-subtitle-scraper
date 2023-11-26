package poller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
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

// AnimeScheduleResponse 구조체는 전체 응답을 정의합니다.
type AnimeScheduleResponse struct {
	Code string      `json:"code"`
	Data []AnimeInfo `json:"data"`
}

// SubtitleInfo 구조체는 자막 정보를 정의합니다.
type SubtitleInfo struct {
	Episode string `json:"episode"` // 자막 회차
	UpdDt   string `json:"updDt"`   // 자막 업로드 시간
	Website string `json:"website"` // 자막 웹사이트
	Name    string `json:"name"`    // 자막 제작자 이름
}

// SubtitleResponse 구조체는 전체 응답을 정의합니다.
type SubtitleResponse struct {
	Code string         `json:"code"`
	Data []SubtitleInfo `json:"data"`
}

// Poller는 주기적으로 API를 통해서 신작 애니메이션 편성표 정보를 받아옵니다.
// 신작 애니메이션 No로 자막 정보를 수집합니다.
// 애니메이션 자막이 업로드되면 scraper에 자막 수집 요청을 보냅니다.
type Poller struct {
	pollingInterval string // Polling interval
	app             *pocketbase.PocketBase
}

// NewPoller는 Poller를 생성합니다.
func NewPoller(pollingInterval string, app *pocketbase.PocketBase) *Poller {
	return &Poller{
		pollingInterval,
		app,
	}
}

// Run은 Poller를 실행합니다.
func (p *Poller) Run() {
	// 1. 신작 애니메이션 편성표 정보를 받아옵니다.
	animeInfos, err := p.GetNewAnimeSchedule()
	if err != nil {
		log.Printf("failed to get new anime schedule: %v", err)
		return
	}
	log.Printf("AnimeCount: %d", len(animeInfos))
	// 2. 신작 애니메이션 No로 자막 정보를 수집합니다.
	for _, animeInfo := range animeInfos {
		subtitleInfos, err := p.GetNewAnimeSubtitleInfo(animeInfo.AnimeNo)
		if err != nil {
			log.Printf("failed to get new anime subtitle info for Anime[%d]: %v", animeInfo.AnimeNo, err)
			continue // 실패한 경우 다음 애니메이션으로 넘어갑니다.
		}
		log.Printf("Anime[%d]-SubtitleCount: %d", animeInfo.AnimeNo, len(subtitleInfos))

		latestSubtitleInfo, ok := ExtractLatestSubtitleInfo(subtitleInfos)
		if !ok {
			log.Printf("failed to get latest subtitle info for Anime[%d]", animeInfo.AnimeNo)
			continue // 실패한 경우 다음 애니메이션으로 넘어갑니다.
		}
		log.Printf("Anime[%d]-LatestSubtitleInfo: %v", animeInfo.AnimeNo, latestSubtitleInfo)

		// 3. if ok true, 유효한 신작 애니메이션 편성표 정보를 DB에 저장합니다.
		err = p.SaveNewAnimeSchedule(animeInfo)
		if err != nil {
			log.Printf("failed to save new anime schedule for Anime[%d]: %v", animeInfo.AnimeNo, err)
			continue // 실패한 경우 다음 애니메이션으로 넘어갑니다.
		}
		// 4. if ok true, 유효한 신작 애니메이션 자막 정보를 DB에 저장합니다.
		err = p.SaveNewAnimeSubtitleInfo(animeInfo, latestSubtitleInfo)
		if err != nil {
			log.Printf("failed to save new anime subtitle info for Anime[%d]: %v", animeInfo.AnimeNo, err)
			continue // 실패한 경우 다음 애니메이션으로 넘어갑니다.
		}
	}
}

// 신작 애니메이션 편성표 정보를 받아옵니다.
func (p *Poller) GetNewAnimeSchedule() ([]AnimeInfo, error) {
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
		var schedule AnimeScheduleResponse
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

// 신작 애니메이션 No로 자막 정보를 수집합니다.
func (p *Poller) GetNewAnimeSubtitleInfo(animeNo int) ([]SubtitleInfo, error) {
	var subtitleInfos []SubtitleInfo
	reqUrl := fmt.Sprintf("https://api.anissia.net/anime/caption/animeNo/%d", animeNo)
	res, err := http.Get(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get new anime subtitle info: %v", err)
	}
	defer res.Body.Close()

	// Check response status code
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get new anime subtitle info: %v", err)
	}

	// JSON을 파싱합니다.
	var subtitleResponse SubtitleResponse
	err = json.NewDecoder(res.Body).Decode(&subtitleResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}
	subtitleInfos = subtitleResponse.Data

	return subtitleInfos, nil
}

func (p *Poller) SaveNewAnimeSchedule(animeInfo AnimeInfo) error {
	animeInfoCollection, err := p.app.Dao().FindCollectionByNameOrId("anime_info")
	if err != nil {
		return fmt.Errorf("failed to find anime_info collection: %v", err)
	}
	// 이미 DB에 저장된 애니메이션인지 확인합니다.
	record, err := p.app.Dao().FindFirstRecordByData("anime_info", "anime_no", animeInfo.AnimeNo)
	if err != nil {
		log.Printf("failed to find anime_info record: %v", err)
		// DB에 저장된 애니메이션이 없는 경우, 새로운 애니메이션 정보를 저장합니다.
		record = models.NewRecord(animeInfoCollection)
		log.Println("New anime_info record created")
	}

	form := forms.NewRecordUpsert(p.app, record)
	form.LoadData(map[string]any{
		"week":          animeInfo.Week,
		"anime_no":      animeInfo.AnimeNo,
		"status":        animeInfo.Status,
		"time":          animeInfo.Time,
		"subject":       animeInfo.Subject,
		"genres":        animeInfo.Genres,
		"caption_count": animeInfo.CaptionCount,
		"start_date":    animeInfo.StartDate,
		"end_date":      animeInfo.EndDate,
		"website":       animeInfo.Website,
	})

	if err := form.Submit(); err != nil {
		return fmt.Errorf("failed to submit form: %v", err)
	}

	return nil
}

func (p *Poller) SaveNewAnimeSubtitleInfo(animeInfo AnimeInfo, subtitleInfo SubtitleInfo) error {
	animeSubtitleCollection, err := p.app.Dao().FindCollectionByNameOrId("anime_subtitle")
	if err != nil {
		return fmt.Errorf("failed to find anime_subtitle collection: %v", err)
	}
	// 이미 DB에 저장된 자막인지 확인합니다.
	// retrieve a single "anime_subtitle" collection record by a string filter expression
	// (use "{:placeholder}" to safely bind untrusted user input parameters)
	record, err := p.app.Dao().FindFirstRecordByFilter(
		"anime_subtitle",
		"anime_no = {:anime_no} && episode = {:episode}",
		dbx.Params{
			"anime_no": animeInfo.AnimeNo,
			"episode":  subtitleInfo.Episode,
		},
	)
	if err != nil {
		log.Printf("failed to find anime_subtitle record: %v", err)
		// DB에 저장된 자막이 없는 경우, 새로운 자막 정보를 저장합니다.
		record = models.NewRecord(animeSubtitleCollection)
		log.Println("New anime_subtitle record created")
	}

	form := forms.NewRecordUpsert(p.app, record)
	form.LoadData(map[string]any{
		"anime_no": animeInfo.AnimeNo,
		"subject":  animeInfo.Subject,
		"episode":  subtitleInfo.Episode,
		"name":     subtitleInfo.Name,
		"website":  subtitleInfo.Website,
	})

	if err := form.Submit(); err != nil {
		return fmt.Errorf("failed to submit form: %v", err)
	}

	return nil
}
