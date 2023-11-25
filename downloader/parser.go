package downloader

import (
	"errors"
	"net/url"
	"strings"
)

// ParserImpl은 downloader.Parser interface 의 구현체입니다.
type ParserImpl struct {
}

// 다운로드 URL 타입을 정의
type DownloadURLType string

const (
	// GoogleDriveURL은 구글 드라이브 URL을 나타냅니다.
	GoogleDriveURL DownloadURLType = "GoogleDrive"
	// NaverBlogURL은 네이버 블로그 URL을 나타냅니다.
	NaverBlogURL DownloadURLType = "NaverBlog"
	// NotSupportedURL은 지원하지 않는 URL을 나타냅니다.
	NotSupportedURL DownloadURLType = "NotSupported"
)

// GetDownloadURLType은 다운로드 URL의 타입을 판별합니다.
func (p *ParserImpl) GetDownloadURLType(url string) DownloadURLType {
	if strings.Contains(url, "https://drive.google.com") {
		return GoogleDriveURL
	} else if strings.Contains(url, "https://download.blog.naver.com/") {
		return NaverBlogURL
	}
	return NotSupportedURL
}

// ParseGoogleDriveURL은 구글 드라이브 URL을 파싱하여 파일 ID를 반환합니다.
func (p *ParserImpl) ParseGoogleDriveURL(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", errors.New("failed to parse url")
	}

	// Google Drive 파일 링크의 경로 부분을 확인
	switch {
	case strings.HasPrefix(parsedURL.Path, "/file/d/"):
		// 링크 형태: https://drive.google.com/file/d/FILE_ID/view?usp=sharing
		parts := strings.Split(parsedURL.Path, "/")
		if len(parts) >= 3 {
			return parts[3], nil
		}
	case strings.HasPrefix(parsedURL.Path, "/uc"):
		// 링크 형태: https://drive.google.com/uc?authuser=0&id=FILE_ID&export=download
		queryParams := parsedURL.Query()
		if id, ok := queryParams["id"]; ok && len(id) > 0 {
			return id[0], nil
		}
	}

	return "", errors.New("invalid Google Drive URL format")
}
