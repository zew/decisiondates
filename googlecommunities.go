package main

import (
	"fmt"
	"html/template"
	"io/ioutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	cus "google.golang.org/api/customsearch/v1"

	"github.com/kataras/iris"

	"github.com/zew/assessmentratedate/config"
	"github.com/zew/assessmentratedate/gorpx"
	"github.com/zew/assessmentratedate/logx"
	"github.com/zew/assessmentratedate/mdl"
	"github.com/zew/assessmentratedate/util"
)

func customSearchService() (*cus.Service, error) {

	// Alternative way to get a client;
	// requires env GOOGLE_APPLICATION_CREDENTIALS=./app_service_account.json
	// Does *not* yield a custom search client.
	if false {
		client, err := google.DefaultClient(oauth2.NoContext)
		_, _ = client, err
	}

	//Get the config from the json key file with the correct scope
	// data, err := ioutil.ReadFile("app_service_account_lib_islands.json")
	data, err := ioutil.ReadFile(config.Config.CredentialFileName)
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

	cses, err := cus.New(client)
	if err != nil {
		fmt.Printf("#3\t%v", err)
		return nil, err
	}

	return cses, nil

}

func results(c *iris.Context) {

	var err error
	display := ""
	respBytes := []byte{}
	strUrl := ""

	if util.EffectiveParam(c, "submit", "none") != "none" {

		start := util.EffectiveParamInt(c, "Start", 1)
		end := util.EffectiveParamInt(c, "Start", 1) + util.EffectiveParamInt(c, "Count", 5)

		//
		//
		communities := []mdl.Community{}
		/*
			      community_id
			    , community_key
				, cleansed as community_name

		*/
		sql := `SELECT 
						*
			FROM 			` + gorpx.TableName(mdl.Community{}) + ` t1
			WHERE 			1=1
				AND		community_id >= :start_id
				AND		community_id <  :end_id
			`
		args := map[string]interface{}{
			"start_id": start,
			"end_id":   end,
		}
		_, err = gorpx.DBMap().Select(&communities, sql, args)
		util.CheckErr(err)
		for i := 0; i < len(communities); i++ {
			logx.Printf("%-4v  %-5v  %v\n", i, communities[i].Id, communities[i].Name)
		}

		cseService, err := customSearchService()
		if err != nil {
			c.Text(200, err.Error())
			return
		}

		for i := 0; i < len(communities); i++ {

			display += fmt.Sprintf("============================\n")
			display += fmt.Sprintf("%v\n", communities[i].Name)

			// https://godoc.org/google.golang.org/api/customsearch/v1
			// CSE Limits you to 10 pages of results with max 10 results per page

			search := cseService.Cse.List(communities[i].Name)
			search.Cx("000184963688878042004:kcoarvtcg7q")
			// search.ExactTerms(communities[i].Key)
			search.ExcludeTerms("factfish")
			search.OrTerms("hebesätze hebesatz")
			search.FileType("pdf")
			search.Safe("off")
			start := int64(1)
			offset := int64(10)     // max allowed is 10
			maxResults := int64(30) // consuming up to three requests; sometimes only 4 results exist

			search.Start(start)
			search.Num(offset)

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
					pdf.CommunityKey = communities[i].Key
					pdf.CommunityName = communities[i].Name
					pdf.Url = r.Link
					pdf.Title = r.Title
					pdf.ResultRank = int(start) + index
					pdf.SnippetGoogle = r.Snippet
					err = gorpx.DBMap().Insert(&pdf)
					if err != nil {
						c.Text(200, err.Error())
					}
				}
				start = start + offset
				// No more search results?
				if start > call.SearchInformation.TotalResults {
					break
				}
			}

		}
	}

	{
		sql := `
			/* update frequencies of pdf urls*/
			UPDATE
				pdf t1
				INNER JOIN (
					SELECT pdf_url, count(*) anz 
					FROM pdf t2
					GROUP BY pdf_url
			) t2 USING (pdf_url)
			SET t1.pdf_frequency = t2.anz
      `
		args := map[string]interface{}{}
		updateRes, err := gorpx.DBMap().Exec(sql, args)
		util.CheckErr(err)
		logx.Printf("updated frequencies: %+v\n", updateRes)

	}

	{
		sql := `
			/* remove pdf_text and snippets for noisy pdfs */
			UPDATE pdf
			SET pdf_snippet1= '', pdf_snippet2= '', pdf_snippet3= ''
			WHERE pdf_frequency > 2
      `
		args := map[string]interface{}{}
		updateRes, err := gorpx.DBMap().Exec(sql, args)
		util.CheckErr(err)
		logx.Printf("emptied : %+v\n", updateRes)

	}

	s := struct {
		HTMLTitle string
		Title     string
		Links     []struct{ Title, Url string }

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
		HTMLTitle: AppName() + " search for pdf docs on each community",
		Title:     AppName() + " search for pdf docs on each community",
		Links:     links,

		StructDump: template.HTML(display),
		RespBytes:  template.HTML(string(respBytes)),

		Url:        strUrl,
		UrlCmp:     "https://www.googleapis.com/customsearch/v1?q=Schwetzingen&key=AIzaSyDS56qRpWj3o_xfGqxwbP5oqW9qr72Poww&cx=000184963688878042004:kcoarvtcg7q",
		FormAction: PathCommunityResults,

		Gemeinde:   util.EffectiveParam(c, "Gemeinde", "Schwetzingen"),
		Schluessel: util.EffectiveParam(c, "Schluessel", "08 2 26 084"),
		ParamStart: util.EffectiveParam(c, "Start", "1"),
		ParamCount: util.EffectiveParam(c, "Count", "5"),
	}

	err = c.Render("results.html", s)
	util.CheckErr(err)

}
