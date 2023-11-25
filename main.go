package main

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/cron"

	"github.com/gocolly/colly/v2"
	"github.com/joho/godotenv"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// link의 타입을 정의
type LinkType int // 0: CommonBlog, 1: NaverBlog

const (
	CommonBlog LinkType = iota
	NaverBlog
)

// link의 타입을 판별한다.
func getLinkType(link string) LinkType {
	if strings.Contains(link, "https://blog.naver.com/") {
		return NaverBlog
	}
	return CommonBlog
}

// a 태그의 href 속성이 https://drive.google.com 인 링크를 전부 찾는다.
func findGoogleDriveLink(url string) ([]string, error) {
	var links []string
	c := colly.NewCollector()

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Check if the link contains the Google Drive prefix
		if strings.HasPrefix(link, "https://drive.google.com") {
			links = append(links, link)
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping
	err := c.Visit(url)
	if err != nil {
		return nil, err
	}

	return links, nil
}

// iframe 태그의 src 속성이 https://download.blog.naver.com 인 링크를 전부 찾는다.
func findNaverBlogLink(url string) ([]string, error) {
	var links []string
	c := colly.NewCollector()

	// iframe의 src 속성을 찾는다
	c.OnHTML("iframe", func(e *colly.HTMLElement) {
		iframeSrc := e.Attr("src")

		// iframe의 src가 존재하면 iframe 내부의 콘텐츠에 대한 크롤링 시작
		if iframeSrc != "" {
			innerCollector := colly.NewCollector()

			innerCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
				link := e.Attr("href")
				if link != "" && e.Request.AbsoluteURL(link) != "" {
					// "https://download.blog.naver.com"으로 시작하는 링크를 출력
					if strings.HasPrefix(e.Request.AbsoluteURL(link), "https://download.blog.naver.com") {
						log.Println("Found link:", e.Request.AbsoluteURL(link))
						links = append(links, e.Request.AbsoluteURL(link))
					}
				}
			})

			innerCollector.OnError(func(r *colly.Response, err error) {
				log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
			})

			// iframe의 src로 요청을 보낸다
			innerCollector.Visit(e.Request.AbsoluteURL(iframeSrc))
		}
	})

	// Start scraping
	err := c.Visit(url)
	if err != nil {
		return nil, err
	}

	return links, nil
}

// 구글 드라이브 파일 링크에서 파일 ID를 추출한다.
func parseGDriveFileID(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
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

func decodeCP949(s string) (string, error) {
	reader := transform.NewReader(strings.NewReader(s), korean.EUCKR.NewDecoder())
	buf := new(strings.Builder)
	_, err := io.Copy(buf, reader)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func decodeFileName(encodedName string) (string, error) {
	// 먼저 UTF-8로 시도
	if utf8.ValidString(encodedName) {
		return encodedName, nil
	}

	// UTF-8이 아니면 CP949로 시도
	decodedName, err := decodeCP949(encodedName)
	if err != nil {
		return "", err
	}

	return decodedName, nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		decodedName, err := decodeFileName(f.Name)
		if err != nil {
			return err
		}

		fPath := filepath.Join(dest, decodedName)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fPath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without overwriting the error
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	// .env 파일을 로드한다.
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}
	// 다운로드 디렉토리를 설정한다.
	downloadDir := os.Getenv("DOWNLOAD_DIR")

	// API 키를 설정한다.
	apiKey := os.Getenv("GDRIVE_API_KEY")
	if apiKey == "" {
		log.Fatalf("failed to get api key")
	}

	// PocketBase를 생성한다.
	app := pocketbase.New()

	// 서버 시작 전에 실행할 함수를 등록한다.
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))

		scheduler := cron.New()

		// call poller every 10 minute
		scheduler.MustAdd("poller", "*/10 * * * *", func() {
			log.Println("Call Poller - Anime Schedule")
		})

		scheduler.Start()

		return nil
	})

	// 서버를 시작한다.
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("failed to create drive service: %v", err)
	}

	links := []string{
		// "https://bluewater91.blogspot.com/2023/11/season-3-7.html",
		// "https://kitauji-highschool.blogspot.com/2023/11/season-3-7.html",
		// "https://ehtelerosa.blogspot.com/2023/10/16bit-another-layer.html",
		// "https://ozuki1.blogspot.com/2023/11/7_17.html",
		// "https://han1sub.blogspot.com/2023/10/blog-post.html",
		"https://blog.naver.com/qtr01122/223268954039",
	}

	for i, link := range links {
		var fileLinks []string

		// 링크 타입을 판별한다.
		linkType := getLinkType(link)
		// 링크 타입에 따라 파일링크를 찾는다.
		switch linkType {
		case CommonBlog:
			// 구글 드라이브 링크를 찾는다.
			log.Println("Common Blog")
			fileLinks, err = findGoogleDriveLink(link)
			if err != nil {
				log.Fatalf("failed to find google drive link: %v", err)
			}
			log.Printf("[%d]Link Count: %d\n", i, len(fileLinks))
			// 파일링크를 순회하며 파일을 다운로드한다.
			for _, fileLink := range fileLinks {
				fileID, err := parseGDriveFileID(fileLink)
				if err != nil {
					log.Fatalf("failed to parse google drive file id: %v", err)
				}

				log.Printf("File ID: %s\n", fileID)

				// 파일 메타데이터를 가져온다.
				file, err := srv.Files.Get(fileID).Do()
				if err != nil {
					log.Fatalf("failed to get file metadata: %v", err)
				}

				log.Printf("File Name: %s\n", file.Name)

				// 파일을 다운로드한다.
				res, err := srv.Files.Get(fileID).Download()
				if err != nil {
					log.Fatalf("failed to download file: %v", err)
				}
				defer res.Body.Close()

				// 다운로드한 파일을 저장한다.
				// 파일 이름은 파일 메타데이터에서 가져온다.
				fileName := strings.TrimSpace(strings.TrimSuffix(file.Name, ".zip"))

				err = os.MkdirAll(fileName, os.ModePerm)
				if err != nil {
					log.Fatalf("failed to create directory: %v", err)
				}

				filePath := fmt.Sprintf("%s/%s/%s", downloadDir, fileName, file.Name)
				f, err := os.Create(filePath)
				if err != nil {
					log.Fatalf("failed to create file: %v", err)
				}
				defer f.Close()

				_, err = io.Copy(f, res.Body)
				if err != nil {
					log.Fatalf("failed to save file: %v", err)
				}

				// zip 파일이면 압축을 푼다.
				if strings.HasSuffix(file.Name, ".zip") {
					unzipPath := fmt.Sprintf("%s/%s", downloadDir, strings.TrimSuffix(file.Name, ".zip"))
					err = unzip(filePath, unzipPath)
					if err != nil {
						log.Fatalf("failed to unzip file: %v", err)
					}
					// 압축 파일은 삭제한다.
					err = os.Remove(filePath)
					if err != nil {
						log.Fatalf("failed to remove zip file: %v", err)
					}
				}
			}
		case NaverBlog:
			// 네이버 블로그 링크를 찾는다.
			log.Println("Naver Blog")
			fileLinks, err = findNaverBlogLink(link)
			if err != nil {
				log.Fatalf("failed to find naver blog link: %v", err)
			}
			log.Printf("[%d]Link Count: %d\n", i, len(fileLinks))
			// TODO: 네이버 블로그 파일 다운로드
			for _, fileLink := range fileLinks {
				resp, err := http.Get(fileLink)
				if err != nil {
					log.Fatalf("failed to download file: %v", err)
				}
				defer resp.Body.Close()

				// Check for correct response code
				if resp.StatusCode != http.StatusOK {
					log.Fatalf("failed to download file: %v", err)
				}

				decodedURL, err := url.QueryUnescape(fileLink)
				if err != nil {
					log.Fatalf("failed to decode url: %v", err)
				}

				log.Printf("File URL: %s\n", decodedURL)

				// 파일 이름을 추출한다.
				fileName := filepath.Base(decodedURL)
				log.Printf("File Name: %s\n", fileName)

				// 파일을 다운로드한다.
				filePath := fmt.Sprintf("%s/%s", downloadDir, fileName)
				f, err := os.Create(filePath)
				if err != nil {
					log.Fatalf("failed to create file: %v", err)
				}
				defer f.Close()

				_, err = io.Copy(f, resp.Body)
				if err != nil {
					log.Fatalf("failed to save file: %v", err)
				}

				// zip 파일이면 압축을 푼다.
				if strings.HasSuffix(fileName, ".zip") {
					unzipPath := fmt.Sprintf("%s/%s", downloadDir, strings.TrimSuffix(fileName, ".zip"))
					err = unzip(filePath, unzipPath)
					if err != nil {
						log.Fatalf("failed to unzip file: %v", err)
					}
					// 압축 파일은 삭제한다.
					err = os.Remove(filePath)
					if err != nil {
						log.Fatalf("failed to remove zip file: %v", err)
					}
				}
			}
		}
	}
}
