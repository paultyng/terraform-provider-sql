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
		StartServer: startContainer(func() (*dockertest.Resource, string, error) {
			resource, err := dockerPool.Run("mysql", "8", []string{
				"MYSQL_ROOT_PASSWORD=tf",
			})
			if err != nil {
				return nil, "", err
			}

			url := fmt.Sprintf("mysql://root:tf@tcp(localhost:%s)/mysql?parseTime=true", resource.GetPort("3306/tcp"))

			return resource, url, nil
		}),

		ExpectedDriver: "mysql",
	},
	{
		ServerType: "postgres",
		StartServer: startContainer(func() (*dockertest.Resource, string, error) {
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
		}),

		ExpectedDriver: "pgx",
	},
	{
		ServerType: "cockroach",
		StartServer: startContainer(func() (*dockertest.Resource, string, error) {
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
		}),
		OnReady: func(db *sql.DB) error {
			_, err := db.Exec("CREATE DATABASE tftest")
			if err != nil {
				return err
			}
			return nil
		},

		ExpectedDriver: "pgx",
	},
	{
		ServerType: "sqlserver",
		StartServer: startContainer(func() (*dockertest.Resource, string, error) {
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
		}),

		ExpectedDriver: "sqlserver",
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

	if testing.Short() {
		return m.Run()
	}

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
		driver.serverOnce = &sync.Once{}
		defer driver.Cleanup()
		go func() error {
			return driver.Start()
		}()
	}

	return m.Run()
}

type testServer struct {
	ServerType  string
	StartServer func() (cleanup func(), url string, err error)
	OnReady     func(*sql.DB) error

	// This is the driver determination expected by the URL scheme
	ExpectedDriver string

	// these are all governed by the sync.Once
	// TODO: support multiple instances, so one test doesn't break
	// another, etc. or maybe just multiple databases in a single server?
	serverOnce    *sync.Once
	url           string
	serverCleanup func()
	serverOnceErr error
}

func (td *testServer) URL() (string, string, error) {
	err := td.Start()
	if err != nil {
		return "", "", err
	}

	scheme, err := schemeFromURL(td.url)

	return td.url, scheme, err
}

func startContainer(fn func() (*dockertest.Resource, string, error)) func() (func(), string, error) {
	return func() (cleanup func(), url string, err error) {
		res, url, err := fn()
		if err != nil {
			return nil, "", err
		}

		// set a hard expiry on the container for 10 minutes
		err = res.Expire(10 * 60)
		if err != nil {
			log.Printf("unable to set hard expiration: %s", err)
			// do not exit here, just log the issue
		}

		err = dockerPool.Retry(func() error {
			p := &provider{}
			err := p.connect(url)
			if err != nil {
				return err
			}
			defer p.DB.Close()

			err = p.DB.Ping()
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return nil, "", err
		}

		return func() {
			if dockerPool != nil {
				dockerPool.Purge(res)
			}
		}, url, nil
	}
}

func (td *testServer) Start() error {
	td.serverOnce.Do(func() {
		td.serverCleanup, td.url, td.serverOnceErr = td.StartServer()
		if td.serverOnceErr != nil {
			return
		}

		if td.OnReady != nil {
			p := &provider{}
			td.serverOnceErr = p.connect(td.url)
			if td.serverOnceErr != nil {
				return
			}
			defer p.DB.Close()

			td.serverOnceErr = td.OnReady(p.DB)
			if td.serverOnceErr != nil {
				return
			}
		}
	})

	return td.serverOnceErr
}

func (td *testServer) Cleanup() {
	if td.serverCleanup != nil {
		td.serverCleanup()
	}
}
