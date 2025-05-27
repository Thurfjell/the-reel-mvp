package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"movez/internal/api/movie"
	"net/http"
	"net/url"
	"time"
)

// Suppose the URLs could be persisted dynamically (eg db) for runtime update.
const genreUrl string = "https://api.themoviedb.org/3/genre/movie/list"
const topListUrl string = "https://api.themoviedb.org/3/discover/movie"
const searchUrl string = "https://api.themoviedb.org/3/search/movie"
const pageQuery string = "page"
const adultQuery string = "include_adult"
const adhocSearchQuery string = "query"
const sortQuery string = "sort_by"
const sortQueryValVoteAvg string = "vote_average.desc"
const voteCountThresholdQuery string = "vote_count.gte"
const voteCountThresholdQueryVal string = "1000"

type Service struct {
	accessToken string
}

type genreApiResponse struct {
	Genres []Genre `json:"genres"`
}

type Genre struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type movieItem struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIDs         []int   `json:"genre_ids"`
	ID               int     `json:"id"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float32 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}

type movieApiResponse struct {
	Page         int         `json:"page"`
	Results      []movieItem `json:"results"`
	TotalPages   int         `json:"total_pages"`
	TotalResults int         `json:"total_results"`
}

type ListItem struct {
	TitleEn     string
	Overview    string
	VoteAverage float32
	GenreIds    []int
	PosterSrc   string
	ReleaseDate time.Time
}

func NewSevice(accessToken string) (s *Service) {
	s = &Service{
		accessToken: accessToken,
	}
	return
}

func (s *Service) GetGenres(ctx context.Context) (genres []movie.Genre, err error) {
	baseUrl, err := url.Parse(genreUrl)
	if err != nil {
		log.Fatal(err)
	}
	params := url.Values{}
	params.Add("language", "en")

	baseUrl.RawQuery = params.Encode()

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(reqCtx, "GET", baseUrl.String(), nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = errors.New("failed to get genres")
		return
	}

	var apiResp genreApiResponse
	if err = json.NewDecoder(res.Body).Decode(&apiResp); err != nil {

		return
	}

	genres = make([]movie.Genre, 0, len(apiResp.Genres))

	for _, g := range apiResp.Genres {
		genres = append(genres, movie.Genre(g))
	}

	return
}

func (s *Service) GetMovies(ctx context.Context, search string) (movies []movie.ListItem, err error) {

	urlToUse := topListUrl
	if len(search) > 0 {
		urlToUse = searchUrl
	}

	baseUrl, err := url.Parse(urlToUse)
	if err != nil {
		log.Fatal(err)
	}
	params := url.Values{}
	params.Add(pageQuery, "1")
	params.Add(adultQuery, "false")
	params.Add(sortQuery, sortQueryValVoteAvg)
	params.Add(voteCountThresholdQuery, voteCountThresholdQueryVal)

	if len(search) > 0 {
		params.Add(adhocSearchQuery, search)
	}

	baseUrl.RawQuery = params.Encode()

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, "GET", baseUrl.String(), nil)

	if err != nil {
		return
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return
	}

	defer res.Body.Close()

	var apiResp movieApiResponse
	if err = json.NewDecoder(res.Body).Decode(&apiResp); err != nil {
		return
	}

	movies = make([]movie.ListItem, 0, len(apiResp.Results))

	for _, m := range apiResp.Results {

		releaseDate, err := parseTime(m.ReleaseDate)
		if err != nil {
			log.Printf("failed to parse date '%s' for movie id %d\n", m.ReleaseDate, m.ID)
			continue
		}

		movies = append(movies, movie.ListItem{
			TitleEn:     m.Title,
			Overview:    m.Overview,
			VoteAverage: m.VoteAverage,
			Genres:      m.GenreIDs,
			PosterSrc:   m.PosterPath,
			ReleaseYear: releaseDate.Year(),
		})
	}

	return
}
