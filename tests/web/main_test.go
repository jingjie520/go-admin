package web

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"testing"

	_ "github.com/GoAdminGroup/go-admin/adapter/gin"
	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/mysql"
	_ "github.com/GoAdminGroup/themes/adminlte"

	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/plugins/admin"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/chartjs"
	"github.com/GoAdminGroup/go-admin/tests/tables"
	"github.com/gin-gonic/gin"
	"github.com/sclevine/agouti"
)

var (
	driver     *agouti.WebDriver
	page       *agouti.Page
	optionList = []string{
		"--headless",
		"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36",
		"--window-size=1000,900",
		"--incognito",
		"--blink-settings=imagesEnabled=true",
		"--no-default-browser-check",
		"--ignore-ssl-errors=true",
		"--ssl-protocol=any",
		"--no-sandbox",
		"--disable-breakpad",
		"--disable-gpu",
		"--disable-logging",
		"--no-zygote",
		"--allow-running-insecure-content",
	}
	quit    = make(chan uint8)
	baseURL = "http://localhost:9033"
	port    = ":9033"
)

func startServer() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	r := gin.New()

	eng := engine.Default()

	adminPlugin := admin.NewAdmin(tables.Generators)
	adminPlugin.AddGenerator("user", tables.GetUserTable)

	template.AddComp(chartjs.NewChart())

	if err := eng.AddConfigFromJSON(os.Args[len(os.Args)-1]).
		AddPlugins(adminPlugin).
		Use(r); err != nil {
		panic(err)
	}

	eng.HTML("GET", "/admin", tables.GetContent)

	r.Static("/uploads", "./uploads")

	go func() {
		_ = r.Run(port)
	}()

	<-quit
	log.Print("closing database connection")
	eng.MysqlConnection().Close()
}

func TestMain(m *testing.M) {

	go startServer()

	var err error

	driver = agouti.ChromeDriver(
		agouti.ChromeOptions("args", optionList),
		agouti.Desired(
			agouti.Capabilities{
				"loggingPrefs": map[string]string{
					"performance": "ALL",
				},
				"acceptSslCerts":      true,
				"acceptInsecureCerts": true,
			},
		))
	err = driver.Start()
	if err != nil {
		panic("failed to start driver, error: " + err.Error())
	}
	defer func() {
		_ = driver.Stop()
	}()

	page, err = driver.NewPage()
	if err != nil {
		panic("failed to open page, error: " + err.Error())
	}
	defer func() {
		_ = page.Destroy()
	}()

	test := m.Run()

	quit <- 0
	os.Exit(test)
}

func StopDriverOnPanic(t *testing.T) {
	if r := recover(); r != nil {
		debug.PrintStack()
		fmt.Println("Recovered in f", r)
		_ = page.Destroy()
		_ = driver.Stop()
		t.Fail()
		quit <- 0
	}
}

func url(suffix string) string {
	return baseURL + suffix
}
