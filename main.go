package main

import (
	"html/template"
	"log"
	"strings"

	"github.com/kataras/iris"
	"github.com/kataras/iris/config"

	appcfg "github.com/zew/assessmentratedate/config"
	"github.com/zew/assessmentratedate/gorpx"
	"github.com/zew/assessmentratedate/logx"
)

var funcMapAll = map[string]interface{}{
	"pref":  Pref,
	"title": strings.Title,
	"toJS":  func(arg string) template.JS { return template.JS(arg) },
}

var irisConfig = config.Iris{}

const (
	PathCommunityResults = "/community-search-results"
	PathProcessPdfs      = "/process-pdfs"
	RefineTextMultiPass  = "/refine-text-multi-pass"
)

const maxPages = 300 // for large Pdf files: Ignore pages greater than
const showMaxXDates = 4
const maxFrequency = 5

var links = []struct{ Title, Url string }{
	{"Search Results per Community", PathCommunityResults},
	{"Extract Text from PDF Files", PathProcessPdfs},
	{"Refine Text Multipass", RefineTextMultiPass},
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

	// iris.Templates("./*.html")

	log.SetFlags(log.Lshortfile)

	var renderOptions = config.Template{
		Directory:  "templates",
		Extensions: []string{".tmpl", ".html"},
		// RequirePartials: true,
		HTMLTemplate: config.HTMLTemplate{
			Funcs: funcMapAll,
		},
	}

	irisConfig.Render.Template = renderOptions
	irisConfig.Render.Template.Layout = "layout.html"

	i01 := iris.New(irisConfig)
	// i01 := iris.Custom(iris.StationOptions{})

	i01.Static(Pref("/js"), "./static/js/", 2)
	// i01.Static("/js", "./static/js/", 1)
	i01.Static(Pref("/img"), "./static/img/", 2)
	i01.Static(Pref("/css"), "./static/css/", 2)

	i01.Get("/", index)
	i01.Get(Pref(""), index)
	i01.Get(Pref("/"), index)

	i01.Get(Pref(PathCommunityResults), results)
	i01.Get(Pref(PathProcessPdfs), processPdf)
	// i01.Get(Pref(PathProcessText), processText)
	// i01.Get(Pref(RefineText), refineText)
	i01.Get(Pref(RefineTextMultiPass), refineTextMultiPass)

	logx.Printf("setting up sql server...")
	gorpx.DBMap()
	defer gorpx.DB().Close()

	logx.Printf("starting http server...")
	logx.Fatal(i01.ListenWithErr(":8082"))

}
