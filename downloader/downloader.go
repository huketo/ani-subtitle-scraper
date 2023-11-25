package downloader

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Unpacker는 압축 파일을 풉니다.
type Unpacker interface {
	// Unpack는 압축 파일을 풉니다.
	Unpack(filePath string, unpackPath string) error
}

// Parser는 다운로드 URL을 파싱합니다.
type Parser interface {
	// GetDownloadURLType은 다운로드 URL의 타입을 판별합니다.
	GetDownloadURLType(url string) DownloadURLType
	// ParseGoogleDriveURL은 구글 드라이브 URL을 파싱하여 파일 ID를 반환합니다.
	ParseGoogleDriveURL(url string) (string, error)
}

// Downloader는 다운로드를 수행합니다.
type Downloader struct {
	GDriveClient *drive.Service // Google Drive API Client
	Parser       Parser         // 다운로드 URL 파서
	DownloadDir  string         // 다운로드 디렉토리
}

// NewDownloader는 Downloader를 생성합니다.
func NewDownloader(ctx context.Context, apiKey string, downloadDir string) *Downloader {
	gdriveClient, err := drive.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("failed to create drive service: %v", err)
	}

	return &Downloader{
		GDriveClient: gdriveClient,
		Parser:       &ParserImpl{},
		DownloadDir:  downloadDir,
	}
}

// Download는 다운로드를 수행합니다.
func (d *Downloader) Download(fileUrl string) error {
	// 다운로드 URL 타입을 판별합니다.
	urlType := d.Parser.GetDownloadURLType(fileUrl)
	log.Printf("URL Type: %s\n", urlType)
	// 다운로드 URL 타입에 따라 다운로드를 수행합니다.
	switch urlType {
	case GoogleDriveURL:
		fileID, err := d.Parser.ParseGoogleDriveURL(fileUrl)
		if err != nil {
			return err
		}
		log.Printf("File ID: %s\n", fileID)

		// 파일 메타데이터를 가져옵니다.
		file, err := d.GDriveClient.Files.Get(fileID).Do()
		if err != nil {
			return err
		}
		log.Printf("File Name: %s\n", file.Name)

		// 파일을 다운로드합니다.
		res, err := d.GDriveClient.Files.Get(fileID).Download()
		if err != nil {
			return err
		}
		defer res.Body.Close()

		// TODO: 파일을 저장합니다.
		// fileName := strings.TrimSpace(strings.TrimSuffix(file.Name, ".zip"))
		// filePath := fmt.Sprintf("%s/%s/%s", downloadDir, fileName, file.Name)

		// saveFile(fileName, res.Body)
	case NaverBlogURL:
		// http 요청을 보냅니다.
		resp, err := http.Get(fileUrl)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// response code를 확인합니다.
		if resp.StatusCode != http.StatusOK {
			return err
		}
		// url을 decode합니다.
		decodedURL, err := url.QueryUnescape(fileUrl)
		if err != nil {
			return err
		}
		fileName := filepath.Base(decodedURL)
		log.Printf("File Name: %s\n", fileName)
	}

	return nil
}
