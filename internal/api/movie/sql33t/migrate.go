package sql33t

import (
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func Up(sqliteName string) (err error) {
	conn, err := sqlite.OpenConn(sqliteName)

	if err != nil {
		return
	}

	// TODO proper err defers all similar defers (-_-) zzz...
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			err = errors.Join(err, closeErr)
		}
		log.Printf("migrations for db %s up done\n", sqliteName)
	}()

	err = sqlitex.Execute(conn, "create table if not exists migrations(version int not null)", &sqlitex.ExecOptions{})
	if err != nil {
		return
	}

	transacFn := sqlitex.Transaction(conn)
	defer transacFn(&err)

	var migratedVersion int
	err = sqlitex.ExecuteTransient(conn, "select version from migrations order by version desc limit 1", &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			migratedVersion = stmt.ColumnInt(0)
			return nil
		},
	})

	dirs, err := os.ReadDir("internal/api/movie/sql33t/migrations")
	type migrationMeta struct {
		Version int
		Name    string
	}
	migrations := make([]migrationMeta, 0)

	for _, dir := range dirs {
		name := []byte(dir.Name())
		versionCutoffIndex := 0
		for i, b := range name {
			if b == []byte(".")[0] {
				versionCutoffIndex = i
				break
			}
		}

		version, _err := strconv.Atoi(string(name[:versionCutoffIndex]))

		if _err != nil {
			err = _err
			return
		}

		if version <= migratedVersion {
			continue
		}

		migrations = append(migrations, migrationMeta{Version: version, Name: dir.Name()})
	}

	slices.SortFunc(migrations, func(a, b migrationMeta) int {
		if a.Version < b.Version {
			return -1
		}

		if a.Version > b.Version {
			return 1
		}

		return 0
	})

	for _, m := range migrations {
		sqlFileBody, _err := os.ReadFile(fmt.Sprintf("internal/api/movie/sql33t/migrations/%s", m.Name))
		if err != nil {
			err = _err
			return
		}
		commands := make([]string, 0)
		s := 0
		for i, b := range sqlFileBody {
			if b == 59 {
				commands = append(commands, string(sqlFileBody[s:i]))
				s = i + 1
			}
		}

		if len(commands) == 0 {
			continue
		}

		for _, c := range commands {
			err = sqlitex.ExecuteTransient(conn, c, &sqlitex.ExecOptions{})
		}
	}

	if err != nil {
		return
	}

	if len(migrations) > 0 {
		err = sqlitex.ExecuteTransient(conn, "insert into migrations(version)values(?)", &sqlitex.ExecOptions{
			Args: []any{migrations[len(migrations)-1].Version},
		})
	}

	return
}
