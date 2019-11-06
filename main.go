package main

import (
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"

	appcfg "github.com/zew/decisiondates/config"
	"github.com/zew/decisiondates/gorpx"
	"github.com/zew/logx"
)

var funcMapAll = map[string]interface{}{
	"pref":  Pref,
	"title": strings.Title,
	"toJS":  func(arg string) template.JS { return template.JS(arg) },
}

const (
	PathCommunityResults = "/community-search-results"
	PathProcessPdfs      = "/process-pdfs"
	RefineTextMultiPass  = "/refine-text-multi-pass"
	DecisionDateEdit     = "/decision-date-edit"
	DecisionDateSave     = "/decision-date-save"
)

const maxPages = 300 // for large Pdf files: Ignore pages greater than
const showMaxXDates = 4
const maxFrequency = 10

var links = []struct{ Title, Url string }{
	{"Search Results per Community", PathCommunityResults},
	{"Extract Text from PDF Files", PathProcessPdfs},
	{"Refine Text Multipass", RefineTextMultiPass},
	{"Update Decision Date", DecisionDateEdit},
}

// The url path prefix
func Pref(p ...string) string {
	s := appcfg.Config.AppName
	s = strings.ToLower(s)
	s = strings.Replace(s, " ", "_", -1)
	if len(p) > 0 {
		return "/" + s + p[0]
	}
	return "/" + s
}

// The name of the application
func AppName(p ...string) string {
	s := appcfg.Config.AppName
	if len(p) > 0 {
		return s + p[0]
	}
	return s
}

func main() {
	log.SetFlags(log.Lshortfile)

	i01 := iris.New()

	view := iris.HTML("./templates", ".html").Layout("layout.html")
	for k, v := range funcMapAll {
		view.AddFunc(k, v)
	}
	i01.RegisterView(view)

	// Used to get and set flash messages on 'decisionedit'.
	sessManager := sessions.New(sessions.Config{
		Cookie:  "sessionid",
		Expires: 2 * time.Hour,
	})
	i01.Use(sessManager.Handler())

	i01.HandleDir(Pref("/js"), "./static/js/")
	// i01.Static("/js", "./static/js/", 1)
	i01.HandleDir(Pref("/img"), "./static/img/")
	i01.HandleDir(Pref("/css"), "./static/css/")

	i01.Get("/", index)
	i01.Get(Pref(""), index)
	i01.Get(Pref("/"), index)

	i01.Get(Pref(PathCommunityResults), results)
	i01.Get(Pref(PathProcessPdfs), processPdf)
	i01.Get(Pref(RefineTextMultiPass), refineTextMultiPass)
	i01.Get(Pref(DecisionDateEdit), decisionDateEdit)
	i01.Post(Pref(DecisionDateSave), decisionDateSave)

	logx.Printf("setting up sql server...")
	gorpx.DBMap()
	defer gorpx.DB().Close()

	logx.Printf("starting http server...")
	logx.Fatal(i01.Run(iris.Addr(":8082"), iris.WithoutServerError(iris.ErrServerClosed)))
}
