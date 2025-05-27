package movie

import (
	"context"
	"log"
)

type ListItem struct {
	TitleEn     string
	Overview    string
	VoteAverage float32
	Genres      []int
	PosterSrc   string
	ReleaseYear int // should probably be time.Time
}

type Genre struct {
	Id   int
	Name string
}

type cache interface {
	GetGenres(ctx context.Context) (genres []Genre, err error)
	SetGenres(ctx context.Context, genres []Genre) (err error)
	GetMovies(ctx context.Context) (movies []ListItem, err error)
	SetMovies(ctx context.Context, movies []ListItem) (err error)
}

type service interface {
	GetGenres(ctx context.Context) (genres []Genre, err error)
	GetMovies(ctx context.Context, search string) (movies []ListItem, err error)
}

type cachedMovieService struct {
	c    cache
	http service
}

func (s *cachedMovieService) GetGenres(ctx context.Context) (genres []Genre, err error) {

	cachedGenres, genresErr := s.c.GetGenres(ctx)

	if genresErr != nil {
		// log alert somewhere.. not fatal
		log.Println("cache err\n", genresErr.Error())
	}

	if len(cachedGenres) > 0 {
		genres = cachedGenres
	}

	if len(genres) == 0 {
		httpGenres, _err := s.http.GetGenres(ctx)
		if _err != nil {
			err = _err
			return
		}

		genres = make([]Genre, 0, len(httpGenres))

		for _, g := range httpGenres {
			genres = append(genres, Genre(g))
		}

		s.c.SetGenres(ctx, genres)

	}
	return
}

func (s *cachedMovieService) GetMovies(ctx context.Context, search string) (movies []ListItem, err error) {
	if len(search) == 0 {
		cachedMovies, cacheErr := s.c.GetMovies(ctx)

		if cacheErr != nil {
			// log alert somewhere.. not fatal
			log.Println("cache err", cacheErr.Error())
		}

		if len(cachedMovies) > 0 {
			movies = cachedMovies
		}

		if len(movies) == 0 {
			serviceMovies, _err := s.http.GetMovies(ctx, search)
			if _err != nil {
				err = _err
				return
			}

			movies = serviceMovies
			s.c.SetMovies(ctx, serviceMovies)
		}
		return
	}

	movies, err = s.http.GetMovies(ctx, search)

	if err != nil {
		return
	}

	return
}

func NewCachedMovieService(svs service, c cache) *cachedMovieService {
	if c == nil {
		log.Fatal("NewCachedMovieService: no cache provided")
	}

	if svs == nil {
		log.Fatal("NewCachedMovieService: no movies service provided")
	}
	return &cachedMovieService{
		c:    c,
		http: svs,
	}
}
