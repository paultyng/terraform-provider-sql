package provider

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/ory/dockertest/v3"
)

var rawServerNames = flag.String("server-types", "mysql,postgres,cockroach,sqlserver", "list of server types to run")

var testServers = []*testServer{
	{
		ServerType: "mysql",
		StartContainer: func() (*dockertest.Resource, string, error) {
			resource, err := dockerPool.Run("mysql", "8", []string{
				"MYSQL_ROOT_PASSWORD=tf",
			})
			if err != nil {
				return nil, "", err
			}

			url := fmt.Sprintf("mysql://root:tf@tcp(localhost:%s)/mysql?parseTime=true", resource.GetPort("3306/tcp"))

			return resource, url, nil
		},
	},
	{
		ServerType: "postgres",
		StartContainer: func() (*dockertest.Resource, string, error) {
			databaseName := "tftest"
			resource, err := dockerPool.Run("postgres", "13", []string{
				"POSTGRES_PASSWORD=tf",
				"POSTGRES_DB=" + databaseName,
				"TZ=UTC",
			})
			if err != nil {
				return nil, "", err
			}

			url := fmt.Sprintf("postgres://postgres:tf@localhost:%s/%s?sslmode=disable", resource.GetPort("5432/tcp"), databaseName)

			return resource, url, nil
		},
	},
	{
		ServerType: "cockroach",
		StartContainer: func() (*dockertest.Resource, string, error) {
			resource, err := dockerPool.RunWithOptions(&dockertest.RunOptions{
				Repository: "cockroachdb/cockroach",
				Tag:        "v20.2.0",
				Cmd: []string{
					"start-single-node",
					"--insecure",
				},
			})
			if err != nil {
				return nil, "", err
			}

			url := fmt.Sprintf("postgres://root@localhost:%s/tftest?sslmode=disable", resource.GetPort("26257/tcp"))

			return resource, url, nil
		},
		OnReady: func(db *sql.DB) error {
			_, err := db.Exec("CREATE DATABASE tftest")
			if err != nil {
				return err
			}
			return nil
		},
	},
	{
		ServerType: "sqlserver",
		StartContainer: func() (*dockertest.Resource, string, error) {
			password := "TF-8chars"
			resource, err := dockerPool.Run("mcr.microsoft.com/mssql/server", "2017-latest", []string{
				"ACCEPT_EULA=y",
				"SA_PASSWORD=" + password,
			})
			if err != nil {
				return nil, "", err
			}

			url := fmt.Sprintf("sqlserver://sa:%s@localhost:%s", password, resource.GetPort("1433/tcp"))

			return resource, url, nil
		},
	},
}

var protoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
	"sql": func() (tfprotov5.ProviderServer, error) {
		return New("acctest")(), nil
	},
}

var dockerPool *dockertest.Pool

func TestMain(m *testing.M) {
	code := runTestMain(m)
	os.Exit(code)
}

func runTestMain(m *testing.M) int {
	var err error

	flag.Parse()

	// remove unspecified test drivers
	serverNames := strings.Split(*rawServerNames, ",")
	for i := len(testServers) - 1; i >= 0; i-- {
		for _, n := range serverNames {
			if strings.TrimSpace(n) == testServers[i].ServerType {
				goto NextServer
			}
		}

		testServers = append(testServers[:i], testServers[i+1:]...)

	NextServer:
	}

	if len(testServers) == 0 {
		log.Fatalf("no test servers specified")
	}

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	poolEndpoint := ""
	if runtime.GOOS == "windows" {
		poolEndpoint = "npipe:////./pipe/docker_engine"
	}
	dockerPool, err = dockertest.NewPool(poolEndpoint)
	if err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}

	dockerPool.MaxWait = 5 * time.Minute

	for _, driver := range testServers {
		driver := driver
		driver.resourceOnce = &sync.Once{}
		defer driver.Cleanup()
		go func() error {
			return driver.Start()
		}()
	}

	return m.Run()
}

type testServer struct {
	ServerType     string
	StartContainer func() (*dockertest.Resource, string, error)
	OnReady        func(*sql.DB) error

	// these are all governed by the sync.Once
	// TODO: support multiple instances, so one test doesn't break
	// another, etc. or maybe just multiple databases in a single server?
	resourceOnce    *sync.Once
	url             string
	resource        *dockertest.Resource
	resourceOnceErr error
}

func (td *testServer) URL() (string, string, error) {
	err := td.Start()
	if err != nil {
		return "", "", err
	}

	scheme, err := schemeFromURL(td.url)

	return td.url, scheme, err
}

func (td *testServer) Start() error {
	td.resourceOnce.Do(func() {
		td.resource, td.url, td.resourceOnceErr = td.StartContainer()
		if td.resourceOnceErr != nil {
			return
		}

		// set a hard expiry on the container for 10 minutes
		err := td.resource.Expire(10 * 60)
		if err != nil {
			log.Printf("unable to set hard expiration: %s", err)
			// do not exit here, just log the issue
		}

		td.resourceOnceErr = dockerPool.Retry(func() error {
			db, err := newDB(td.url, nil)
			if err != nil {
				return err
			}
			defer db.Close()

			err = db.Ping()
			if err != nil {
				return err
			}

			return nil
		})
		if td.resourceOnceErr != nil {
			return
		}

		if td.OnReady != nil {
			var db *db
			db, td.resourceOnceErr = newDB(td.url, nil)
			if td.resourceOnceErr != nil {
				return
			}
			defer db.Close()

			td.resourceOnceErr = td.OnReady(db.DB)
			if td.resourceOnceErr != nil {
				return
			}
		}
	})

	return td.resourceOnceErr
}

func (td *testServer) Cleanup() {
	if dockerPool != nil && td.resource != nil {
		dockerPool.Purge(td.resource)
	}
}
