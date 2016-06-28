package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	cus "google.golang.org/api/customsearch/v1"

	"github.com/kataras/iris"

	"github.com/zew/assessmentratedate/gorpx"
	"github.com/zew/assessmentratedate/logx"
	"github.com/zew/assessmentratedate/mdl"
	"github.com/zew/assessmentratedate/util"
)

func getService() (*cus.Service, error) {

	// requires env GOOGLE_APPLICATION_CREDENTIALS=./app_service_account.json
	// Does not work
	if false {
		client, err := google.DefaultClient(oauth2.NoContext)
		_, _ = client, err
	}

	//Get the config from the json key file with the correct scope
	data, err := ioutil.ReadFile("app_service_account.json")
	if err != nil {
		fmt.Printf("#1\t%v", err)
		return nil, err
	}
	conf, err := google.JWTConfigFromJSON(data, "https://www.googleapis.com/auth/cse")
	if err != nil {
		fmt.Printf("#2\t%v", err)
		return nil, err
	}
	client := conf.Client(oauth2.NoContext)

	customsearchService, err := cus.New(client)
	if err != nil {
		fmt.Printf("#3\t%v", err)
		return nil, err
	}

	return customsearchService, nil

}

func results(c *iris.Context) {

	var err error
	display := ""
	respBytes := []byte{}
	community := mdl.Community{}
	strUrl := ""

	if util.EffectiveParam(c, "submit", "none") != "none" {

		//
		//
		communities := []mdl.Community{}
		sql := `SELECT 
				      community_key
					, cleansed as community_name
			FROM 			` + gorpx.TableName(mdl.Community{}) + ` t1
			WHERE 			1=1
				AND		community_id >= :start_id
				AND		community_id <  :end_id
			`
		args := map[string]interface{}{
			"start_id": util.EffectiveParamInt(c, "Start", 1),
			"end_id":   util.EffectiveParamInt(c, "Start", 1) + util.EffectiveParamInt(c, "Count", 5),
		}

		_, err = gorpx.DBMap().Select(&communities, sql, args)

		//
		logx.Printf("%+v\n", communities)

		if false {

			myUrl := url.URL{}
			myUrl.Host = "www.googleapis.com"
			myUrl.Path = "customsearch/v1"
			myUrl.Scheme = "https"
			logx.Printf("host is %v", myUrl.String())

			// https://developers.google.com/apis-explorer/#p/customsearch/v1/search.cse.list?q=Schwetzingen&_h=1&
			// exactTerms: Identifies a phrase that all documents in the search results must contain (string)
			// excludeTerms
			// fileType: pdf
			// num: Number of search results to return (integer)

			vals := map[string]string{
				"key":   "AIzaSyDS56qRpWj3o_xfGqxwbP5oqW9qr72Poww", // "Server key 1" from libertarian islands
				"cx":    "000184963688878042004:kcoarvtcg7q",       // searchEngineId for google custom search engine "cse"
				"q":     util.EffectiveParam(c, "Gemeinde", "Villingen-Schwenningen"),
				"start": util.EffectiveParam(c, "Start", "1"),
				"num":   util.EffectiveParam(c, "Count", "5"),
				"safe":  "off",
			}

			queryStr := ""
			for k, v := range vals {
				queryStr += fmt.Sprintf("%v=%v&", k, v)
			}
			logx.Printf("queryStr is %v", queryStr)

			strUrl = myUrl.String() + "/?" + queryStr
			req, err := http.NewRequest("GET", strUrl, nil)
			util.CheckErr(err)

			resp, err := util.HttpClient().Do(req)
			util.CheckErr(err)
			defer resp.Body.Close()

			respBytes, err = ioutil.ReadAll(resp.Body)
			util.CheckErr(err)

			// Parse
			if err != nil {
				c.Text(200, err.Error())
				return
			}

			err = gorpx.DBMap().Insert(&community)
			if err != nil {
				c.Text(200, err.Error())
			}

			display = util.IndentedDump(community)
			// c.Text(200, display)

		} else {

			cseService, err := getService()
			if err != nil {
				c.Text(200, err.Error())
				return
			}

			// CSE Limits you to 10 pages of results with max 10 results per page

			start := int64(1)
			offset := int64(3)
			maxResults := int64(8)

			search := cseService.Cse.List(communities[0].Name)
			search.Cx("000184963688878042004:kcoarvtcg7q")
			search.ExcludeTerms("factfish")
			// search.ExactTerms(communities[0].Key)
			search.OrTerms("hebesätze hebesatz")
			search.FileType("pdf")
			search.Start(start)
			search.Num(offset)
			search.Safe("off")

			display += "so far\n"

			for start < maxResults {
				search.Start(int64(start))
				call, err := search.Do()
				if err != nil {
					c.Text(200, err.Error())
					return
				}
				for index, r := range call.Items {
					display += fmt.Sprintf("%-4v %-22v %-32v  %v\n", start+int64(index), r.FileFormat, r.Link, r.DisplayLink)
					display += fmt.Sprintf("%v\n", r.Title)
					display += fmt.Sprintf("%v\n", r.Snippet)
					// display += fmt.Sprintf("%+v\n", r)
					display += fmt.Sprintf("\n")

					pdf := mdl.Pdf{}
					pdf.Key = communities[0].Key
					pdf.Name = communities[0].Name
					pdf.Url = r.Link
					pdf.Snippet1 = fmt.Sprintf("%v\n\n%v", r.Title, r.Snippet)
					err = gorpx.DBMap().Insert(&pdf)
					if err != nil {
						c.Text(200, err.Error())
					}
				}
				start = start + offset
				// No more search results?
				if call.SearchInformation.TotalResults < start {
					break
				}
			}

		}
	}

	s := struct {
		HTMLTitle string
		Title     string
		Links     map[string]string

		FormAction string
		Gemeinde   string
		Schluessel string
		ParamStart string
		ParamCount string

		Url    string
		UrlCmp string

		StructDump template.HTML
		RespBytes  template.HTML
	}{
		HTMLTitle: AppName() + " top sites",
		Title:     AppName() + " top sites",
		Links:     links,

		StructDump: template.HTML(display),
		RespBytes:  template.HTML(string(respBytes)),

		Url:        strUrl,
		UrlCmp:     "https://www.googleapis.com/customsearch/v1?q=Schwetzingen&key=AIzaSyDS56qRpWj3o_xfGqxwbP5oqW9qr72Poww&cx=000184963688878042004:kcoarvtcg7q",
		FormAction: PathResults,

		Gemeinde:   util.EffectiveParam(c, "Gemeinde", "Schwetzingen"),
		Schluessel: util.EffectiveParam(c, "Schluessel", "08 2 26 084"),
		ParamStart: util.EffectiveParam(c, "Start", "1"),
		ParamCount: util.EffectiveParam(c, "Count", "5"),
	}

	err = c.Render("results.html", s)
	util.CheckErr(err)

}
