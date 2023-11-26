package poller

import (
	"log"
	"strconv"
	"time"
)

// ExtractLatestSubtitleInfo 함수는 제공된 자막 정보 슬라이스에서 최신 자막 정보를 추출합니다.
// 만약 website 필드가 비어 있는 경우, false를 반환합니다.
func ExtractLatestSubtitleInfo(subtitleInfos []SubtitleInfo) (SubtitleInfo, bool) {
	var latestSubtitleInfo SubtitleInfo
	if len(subtitleInfos) > 0 {
		maxEpisode, _ := strconv.ParseFloat(subtitleInfos[0].Episode, 64)
		maxUpdDtParsed, err := time.Parse("2006-01-02T15:04:05", subtitleInfos[0].UpdDt)
		if err != nil {
			log.Printf("failed to parse date: %v", err)
			return SubtitleInfo{}, false
		}
		latestSubtitleInfo = subtitleInfos[0]

		for _, subtitleInfo := range subtitleInfos {
			episode, err := strconv.ParseFloat(subtitleInfo.Episode, 64)
			if err != nil {
				log.Printf("failed to parse episode number: %v", err)
				continue
			}

			updDtParsed, err := time.Parse("2006-01-02T15:04:05", subtitleInfo.UpdDt)
			if err != nil {
				log.Printf("failed to parse date: %v", err)
				continue
			}

			if episode > maxEpisode || (episode == maxEpisode && updDtParsed.Before(maxUpdDtParsed)) {
				maxEpisode = episode
				maxUpdDtParsed = updDtParsed
				latestSubtitleInfo = subtitleInfo
			}
		}

		if latestSubtitleInfo.Website != "" {
			return latestSubtitleInfo, true
		}
	}
	return SubtitleInfo{}, false
}
