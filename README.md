# Anime Subtitle Scraper(anisub-scraper)

`anisub-scraper`는 [Anissia API] 를 사용해서 신작 애니메이션 정보와 애니메이션 자막 제작자의 정보를 받아와서 애니메이션 자막을 수집하는 프로그램입니다.

[Anissia API]: https://github.com/anissia-net/document/blob/main/api_anime_schdule.md

## 어떻게 작동하는가?

1. GET `https://api.anissia.net/anime/schedule/{{week}}` 요청으로 신작 애니메이션 편성표 정보를 받아 옵니다.

Sample Schedule Data:

```json
{
  "animeNo": 2508,
  "week": "0",
  "status": "ON",
  "time": "01:00",
  "subject": "티어문 제국 이야기 ~단두대에서 시작하는 황녀님의 전생 역전 스토리~",
  "genres": "판타지",
  "captionCount": 4,
  "startDate": "2023-10-08",
  "endDate": "",
  "website": "https://tearmoon-pr.com/"
}
```

2. `captionCount`가 0보다 큰 데이터를 저장합니다.

3. GET `https://api.anissia.net/anime/caption/animeNo/2508` 요청으로 자막 제작자 정보를 가져옵니다.

Sample Caption Data:

```json
[
  {
    "episode": "7",
    "updDt": "2023-11-19T23:27:00",
    "website": "https://kitauji-highschool.blogspot.com/2023/11/7_21.html",
    "name": "냥키치"
  },
  {
    "episode": "7",
    "updDt": "2023-11-19T03:42:00",
    "website": "https://bluewater91.blogspot.com/2023/11/7_19.html",
    "name": "별명따위"
  },
  {
    "episode": "5",
    "updDt": "2023-11-06T00:31:00",
    "website": "https://blog.naver.com/cndska15/223256605784",
    "name": "코코아레인"
  },
  {
    "episode": "3.1",
    "updDt": "2023-10-22T04:24:00",
    "website": "https://felia.tistory.com/885",
    "name": "코코렛"
  }
]
```

4. 자막정보에서 `episode` 가 가장 높고 `updDt`가 빠른 자막정보를 저장합니다.

### 애니메이션 테이블

- animeNo
- subject
- episode
- subtitles
- createAt
- updateAt

### 자막 테이블

- subNo
- animeNo
- subject
- episode
- subtitle
- createAt
- updateAt

5. html을 파싱하여 다운로드 링크를 찾아낸다.
6. 다운로드 링크의 유형을 분류한다.
7. 자막을 다운로드 받는다.
8. 다운로드 파일에 폰트가 존재할 경우 폰트 파일을 따로 저장한다.

## Build & Run

```bash
CGO_ENABLED=0 go build -o ./build/anisub-scraper
./build/anisub-scraper serve
```
