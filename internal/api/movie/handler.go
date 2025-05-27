package movie

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"math"
	"movez/internal/api"
	"net/http"
	"sync"
)

//go:embed templates/*
var html embed.FS

const moviesHtml string = "movies.html"
const movieListHtml string = "movie_list.html"

type HandlerData interface {
	GetGenres(ctx context.Context) (genres []Genre, err error)
	GetMovies(ctx context.Context, search string) (movies []ListItem, err error)
}

type Handler struct {
	Template *template.Template
	data     HandlerData
	bufPool  *sync.Pool
}

func (h *Handler) Routes() []api.RouteMeta {

	return []api.RouteMeta{
		h.getMovies(),
		h.getMovieList(),
	}
}

func NewHandler(data HandlerData) (handler *Handler, err error) {
	template, err := template.ParseFS(html, "templates/*.html")

	if err != nil {
		return
	}

	handler = &Handler{
		Template: template,
		data:     data,
		bufPool: &sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
	return
}

func (h *Handler) getMovies() api.RouteMeta {
	return api.RouteMeta{
		Path: "GET /movies",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := h.Template.ExecuteTemplate(w, moviesHtml, nil)

			if err != nil {
				w.Write([]byte("SORRY! :D"))
			}
		}),
	}
}

type listItem struct {
	TitleEn     string
	Overview    string
	VoteAverage float64
	Genres      []string
	PosterSrc   string
	ReleaseYear int
}

func (h *Handler) getMovieList() api.RouteMeta {
	return api.RouteMeta{
		Path: "GET /movie-list",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			search := r.URL.Query().Get("search")
			movies, err := h.data.GetMovies(r.Context(), search)

			if err != nil {
				w.Write([]byte("SORRY! :D"))
			}

			genres, err := h.data.GetGenres(r.Context())
			if err != nil {
				w.Write([]byte("SORRY! :D"))
			}

			type model struct {
				Movies []listItem
			}

			m := &model{
				Movies: make([]listItem, 0, len(movies)),
			}

			for _, movie := range movies {
				genreNames := make([]string, 0, len(movie.Genres))
				for _, g := range movie.Genres {
					for _, s := range genres {
						if s.Id == g {
							genreNames = append(genreNames, s.Name)
							break
						}
					}
				}

				m.Movies = append(m.Movies, listItem{
					TitleEn:     movie.TitleEn,
					Genres:      genreNames,
					Overview:    movie.Overview,
					VoteAverage: math.Round(float64(movie.VoteAverage)*10) / 10,
					PosterSrc:   movie.PosterSrc,
					ReleaseYear: movie.ReleaseYear,
				})
			}

			err = h.Template.ExecuteTemplate(w, movieListHtml, m)

			if err != nil {
				w.Write([]byte("SORRY! :D"))
			}
		}),
	}
}
