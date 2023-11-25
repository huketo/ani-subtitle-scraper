package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

// a 태그의 href 속성이 https://drive.google.com 인 링크를 전부 찾는다.
func findGoogleDriveLink(url string) ([]string, error) {
	var links []string

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// HTML 문서 파싱
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	// a 태그를 찾는다.
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			// a 태그의 href 속성을 찾는다.
			for _, a := range n.Attr {
				if a.Key == "href" {
					// https://drive.google.com 인 링크를 찾는다.
					if strings.HasPrefix(a.Val, "https://drive.google.com") {
						links = append(links, a.Val)
					}
				}
			}
		}

		// 재귀적으로 탐색
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return links, nil
}

func checkGoogleDriveLink(url string) bool {
	return strings.HasPrefix(url, "https://drive.google.com/uc?export=download&id=")
}

func parseGoogleDriveLink(url string) (string, error) {
	// 이런 형태의 링크를 파싱하여
	// https://drive.google.com/file/d/1jM-wpj_eMxJMAY11ew8TYMZkLwm8e5_e/view?usp=sharing
	// https://drive.google.com/file/d/1zddjpst2hVqipfQc1aQSvl6pDQNUAMX-/view?usp=drive_link
	// https://drive.google.com/uc?authuser=0&id=1B9775i9zv3d_HNy0FGKtM2xNmL34p46O&export=download

	// 이런 형태로 변환한다.
	// https://drive.google.com/uc?export=download&id=1km2cxblo_kSCpQ-rjD1pzQasmx9HfGam

	// 변환 결과
	// https://drive.google.com/uc?export=download&id=1jM-wpj_eMxJMAY11ew8TYMZkLwm8e5_e

	fileID := ""
	if strings.HasPrefix(url, "https://drive.google.com/file/d/") {
		fileID = strings.TrimPrefix(url, "https://drive.google.com/file/d/")
	} else if strings.HasPrefix(url, "https://drive.google.com/uc?authuser=0&id=") {
		fileID = strings.TrimPrefix(url, "https://drive.google.com/uc?authuser=0&id=")
	} else {
		return "", errors.New("invalid google drive link")
	}

	if strings.Contains(fileID, "/view?usp=sharing") {
		fileID = strings.TrimSuffix(fileID, "/view?usp=sharing")
	} else if strings.Contains(fileID, "/view?usp=drive_link") {
		fileID = strings.TrimSuffix(fileID, "/view?usp=drive_link")
	} else if strings.Contains(fileID, "&export=download") {
		fileID = strings.TrimSuffix(fileID, "&export=download")
	}

	url = "https://drive.google.com/uc?export=download&id=" + fileID

	return url, nil
}

func downloadFile(URL, fileName string) error {
	// Get the data
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for the correct response
	if resp.StatusCode != http.StatusOK {
		return errors.New("received non-200 response code")
	}

	// Create the file
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
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

func downloadGoogleDriveFiles(fileLinks []string) error {
	for i, fileLink := range fileLinks {
		isDownloadable := checkGoogleDriveLink(fileLink)
		fmt.Printf("isDownloadable[%d]: %v\n", i, isDownloadable)

		if !isDownloadable {
			var err error
			fileLink, err = parseGoogleDriveLink(fileLink)
			if err != nil {
				return err
			}
		}

		fmt.Printf("file_link[%d]: %s\n", i, fileLink)

		// download 폴더가 없으면 생성
		if _, err := os.Stat("./download"); os.IsNotExist(err) {
			err = os.Mkdir("./download", os.ModePerm)
			if err != nil {
				return err
			}
		}

		zipFile := fmt.Sprintf("./download/downloaded_%d.zip", i)
		err := downloadFile(fileLink, zipFile)
		if err != nil {
			return err
		}

		// ZIP 파일 압축 해제
		err = unzip(zipFile, "./download")
		if err != nil {
			return err
		}

		// ZIP 파일 삭제
		err = os.Remove(zipFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	links := []string{
		"https://bluewater91.blogspot.com/2023/11/season-3-7.html",
		"https://kitauji-highschool.blogspot.com/2023/11/season-3-7.html",
		"https://ehtelerosa.blogspot.com/2023/10/16bit-another-layer.html",
		"https://ozuki1.blogspot.com/2023/11/7_17.html",
		"https://han1sub.blogspot.com/2023/10/blog-post.html",
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(links))

	for i, link := range links {
		fileLinks, err := findGoogleDriveLink(link)
		if err != nil {
			fmt.Printf("failed to find google drive link: %v", err)
			panic(err)
		}

		fmt.Printf("[%d]Link Count: %d\n", i, len(fileLinks))

		err = downloadGoogleDriveFiles(fileLinks)
		if err != nil {
			fmt.Printf("failed to download google drive files: %v", err)
			panic(err)
		}
	}

	fmt.Println("Done")
}
