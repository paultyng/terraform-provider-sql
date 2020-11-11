package provider

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest"
	"golang.org/x/sync/errgroup"
)

var testURLs = map[string]string{}

var defaultTestFactories = map[string]func(pool *dockertest.Pool) (*dockertest.Resource, func() string, error){
	"mysql":     startMySQL,
	"postgres":  startPostgres,
	"cockroach": startCockroach,
	"sqlserver": startSQLServer,
}

func TestMain(m *testing.M) {
	code := runTestMain(m)
	os.Exit(code)
}

func runTestMain(m *testing.M) int {
	// TODO: move this to sync.once style setup to better support -short

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("npipe:////./pipe/docker_engine")
	if err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}

	pool.MaxWait = 2 * time.Minute

	eg := &errgroup.Group{}

	for k, factory := range defaultTestFactories {
		r, urlF, err := factory(pool)
		if err != nil {
			log.Printf("unable to start container %q: %s", k, err)
			return -1
		}
		defer pool.Purge(r)

		// set a hard expiry on the container for 10 minutes
		err = r.Expire(10 * 60)
		if err != nil {
			log.Printf("unable to set hard expiration: %s", err)
			// do not exit here, just log the issue
		}

		eg.Go(func() error {
			if err := pool.Retry(func() error {
				url := urlF()
				testURLs[k] = url

				db, err := newDB(url, nil)
				if err != nil {
					return err
				}
				defer db.Close()

				err = db.Ping()
				if err != nil {
					return err
				}

				return nil
			}); err != nil {
				return fmt.Errorf("timeout waiting for ping for %q: %w", k, err)
			}

			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		log.Printf("Error starting databases: %s", err)
		return -1
	}

	return m.Run()
}

func startMySQL(pool *dockertest.Pool) (*dockertest.Resource, func() string, error) {
	resource, err := pool.Run("mysql", "8", []string{
		"MYSQL_ROOT_PASSWORD=tf",
	})
	if err != nil {
		return nil, nil, err
	}

	return resource, func() string {
		// TODO: automatically set parseTime here?
		return fmt.Sprintf("mysql://root:tf@tcp(localhost:%s)/mysql?parseTime=true", resource.GetPort("3306/tcp"))
	}, nil
}

func startPostgres(pool *dockertest.Pool) (*dockertest.Resource, func() string, error) {
	databaseName := "tftest"
	resource, err := pool.Run("postgres", "13", []string{
		"POSTGRES_PASSWORD=tf",
		"POSTGRES_DB=" + databaseName,
		"TZ=UTC",
	})
	if err != nil {
		return nil, nil, err
	}

	return resource, func() string {
		return fmt.Sprintf("postgres://postgres:tf@localhost:%s/%s?sslmode=disable", resource.GetPort("5432/tcp"), databaseName)
	}, nil
}

func startCockroach(pool *dockertest.Pool) (*dockertest.Resource, func() string, error) {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "cockroachdb/cockroach",
		Tag:        "v20.2.0",
		Cmd: []string{
			"start-single-node",
			"--insecure",
		},
	})
	if err != nil {
		return nil, nil, err
	}

	return resource, func() string {
		return fmt.Sprintf("postgres://root@localhost:%s/events?sslmode=disable", resource.GetPort("26257/tcp"))
	}, nil
}

func startSQLServer(pool *dockertest.Pool) (*dockertest.Resource, func() string, error) {
	password := "TF-8chars"
	resource, err := pool.Run("mcr.microsoft.com/mssql/server", "2017-latest", []string{
		"ACCEPT_EULA=y",
		"SA_PASSWORD=" + password,
	})
	if err != nil {
		return nil, nil, err
	}

	return resource, func() string {
		return fmt.Sprintf("sqlserver://sa:%s@localhost:%s", password, resource.GetPort("1433/tcp"))
	}, nil
}
