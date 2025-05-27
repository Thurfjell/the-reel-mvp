package sql33t

import (
	"context"
	"errors"
	"log"
	"movez/internal/api/movie"
	"strings"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

// Let's be OK with this sqlite cache depending on movie package since we can imagine it as being a sub package of movies
type cache struct {
	Pool    *sqlitex.Pool
	Timeout time.Duration
}

// (͡° ͜ʖ ͡°)
func NewCache(sqliteName string) (c *cache, err error) {
	pool, err := sqlitex.NewPool(sqliteName, sqlitex.PoolOptions{PoolSize: 200})

	if err != nil {
		return
	}

	c = &cache{
		Pool:    pool,
		Timeout: 5 * time.Second,
	}
	return
}

func (c *cache) GetGenres(ctx context.Context) (genres []movie.Genre, err error) {
	connCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	conn, err := c.Pool.Take(connCtx)
	if err != nil {
		return
	}
	defer c.Pool.Put(conn)

	sqlitex.Execute(conn, "with c as(select count(*) from genres) select id, name, (select * from c) from genres", &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			if len(genres) == 0 {
				genres = make([]movie.Genre, 0, stmt.ColumnInt(2))
			}
			genres = append(genres, movie.Genre{
				Id:   stmt.ColumnInt(0),
				Name: stmt.ColumnText(1),
			})
			return nil
		},
	})

	if genres == nil {
		err = errors.New("no genres in cache")
	}

	return
}

func (c *cache) SetGenres(ctx context.Context, genres []movie.Genre) (err error) {
	connCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	conn, err := c.Pool.Take(connCtx)
	if err != nil {
		return
	}
	defer c.Pool.Put(conn)

	stmt, err := conn.Prepare("insert into genres(id, name) values(?,?)")

	if err != nil {
		return
	}

	defer func() {
		serr := stmt.Finalize()
		if serr != nil {
			err = errors.Join(err, serr)
		}
	}()

	for _, g := range genres {
		err = stmt.ClearBindings()
		err = stmt.Reset()
		stmt.BindInt64(1, int64(g.Id))
		stmt.BindText(2, g.Name)

		if hasRow, err := stmt.Step(); err != nil {
			return err
		} else if hasRow {
			return errors.New("bulk genres returned row during insert")
		}
	}

	return
}

// Crude cache of the default load content
func (c *cache) GetMovies(ctx context.Context) (movies []movie.ListItem, err error) {
	connCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	conn, err := c.Pool.Take(connCtx)
	if err != nil {
		return
	}
	defer c.Pool.Put(conn)

	err = sqlitex.Execute(conn, strings.TrimSpace(`
		with c as(
			select count(*) from movies
		)
		select 
			(select * from c),
			title_en,
			overview,
			vote_avg,
			genre_ids_csv,
			poster_src,
			release_year
		from movies
	`), &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			if len(movies) == 0 {
				movies = make([]movie.ListItem, 0, stmt.ColumnInt(0))
			}

			movies = append(movies, movie.ListItem{
				TitleEn:     stmt.ColumnText(1),
				Overview:    stmt.ColumnText(2),
				VoteAverage: float32(stmt.ColumnFloat(3)),
				Genres:      csvToInts(stmt.ColumnText(4)),
				PosterSrc:   stmt.ColumnText(5),
				ReleaseYear: stmt.ColumnInt(6),
			})
			return nil
		},
	})

	if movies == nil {
		err = errors.New("no movies in cache")
	}

	return
}

func (c *cache) SetMovies(ctx context.Context, movies []movie.ListItem) (err error) {
	connCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	conn, err := c.Pool.Take(connCtx)
	if err != nil {
		return
	}
	defer c.Pool.Put(conn)

	stmt, err := conn.Prepare("insert into movies(overview, poster_src, title_en, genre_ids_csv, release_year, vote_avg) values (?,?,?,?,?,?)")

	if err != nil {
		return
	}

	defer func() {
		serr := stmt.Finalize()
		if serr != nil {
			err = errors.Join(err, serr)
		}
	}()

	for _, m := range movies {
		err = stmt.ClearBindings()
		err = stmt.Reset()

		stmt.BindText(1, m.Overview)
		stmt.BindText(2, m.PosterSrc)
		stmt.BindText(3, m.TitleEn)
		stmt.BindText(4, intsToCsv(m.Genres))
		stmt.BindInt64(5, int64(m.ReleaseYear))
		stmt.BindFloat(6, float64(m.VoteAverage))

		if hasRow, err := stmt.Step(); err != nil {
			return err
		} else if hasRow {
			return errors.New("bulk movies returned row during insert")
		}
	}
	return
}

func (c *cache) Close() (err error) {
	err = c.Pool.Close()
	log.Println("pool closed")
	return
}
