package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/kataras/iris"
	"github.com/rsc.io/pdf"

	"github.com/zew/assessmentratedate/gorpx"
	"github.com/zew/assessmentratedate/logx"
	"github.com/zew/assessmentratedate/mdl"
	"github.com/zew/assessmentratedate/util"
)

func processPdf(c *iris.Context) {

	var err error
	display := ""
	respBytes := []byte{}
	strUrl := ""

	if util.EffectiveParam(c, "submit", "none") != "none" {

		start := util.EffectiveParamInt(c, "Start", 1)
		end := util.EffectiveParamInt(c, "Start", 1) + util.EffectiveParamInt(c, "Count", 5)

		//
		//
		pdfs := []mdl.Pdf{}
		sql := `SELECT 
					/*
				      community_key
					, cleansed as community_name

					*/
					*
			FROM 			` + gorpx.TableName(mdl.Pdf{}) + ` t1
			WHERE 			1=1
				AND		pdf_id >= :start_id
				AND		pdf_id <  :end_id
			`
		args := map[string]interface{}{
			"start_id": start,
			"end_id":   end,
		}
		_, err = gorpx.DBMap().Select(&pdfs, sql, args)
		util.CheckErr(err)
		for i := 0; i < len(pdfs); i++ {
			logx.Printf("%-4v  %55v %v\n", i, util.UpTo(pdfs[i].Url, 70), pdfs[i].Title) // dont print out all fields
		}

		for i := 0; i < len(pdfs); i++ {

			display += fmt.Sprintf("============================\n")
			display += fmt.Sprintf("%v\n", pdfs[i].CommunityName)

			// vals := map[string]string{
			// 	"ResponseGroup": "Country",
			// }
			// queryStr := ""
			// for k, v := range vals {
			// 	queryStr += fmt.Sprintf("%v=%v&", k, v)
			// }

			req, err := http.NewRequest("GET", pdfs[i].Url, nil)
			strUrl = pdfs[i].Url
			util.CheckErr(err)

			resp, err := util.HttpClient().Do(req)
			util.CheckErr(err)
			defer resp.Body.Close()

			respBytes, err = ioutil.ReadAll(resp.Body)
			util.CheckErr(err)

			ioutil.WriteFile(fmt.Sprintf("pdfNumber%03v.pdf", i), respBytes, os.FileMode(777))

			rdr := bytes.NewReader(respBytes)

			logx.Printf("opening pdf #%-2v %v\n", i, util.UpToR(pdfs[i].Url, 70))
			display += fmt.Sprintf("opening pdf #%-2v %v\n", i, util.UpToR(pdfs[i].Url, 70))

			var rdr2 *pdf.Reader
			errFct := func() error {
				var errInner error
				defer func() {
					if r := recover(); r != nil {
						logx.Printf("Recovered while creating reader:", r)
						errInner = fmt.Errorf("Recovered while creating reader:", r)
					}
				}()
				rdr2, errInner = pdf.NewReader(rdr, int64(len(respBytes)))
				return errInner
			}()
			if errFct != nil {
				continue
			}

			content := bytes.Buffer{}

			numPages := rdr2.NumPage()
			logx.Printf("  found %v pages\n", numPages)
			for j := 1; j <= numPages; j++ {
				if j > 6 {
					logx.Printf("  not opening more than 6 pages\n")
					break
				}
				page := rdr2.Page(j)
				content.WriteString(fmt.Sprintf("\nopening page %v\n", j))
				logx.Printf("   opening page %v\n", j)

				cn := page.Content()
				texts := cn.Text
				for k := 0; k < len(texts); k++ {
					if k > 444 {
						// logx.Printf("       not opening more than 444 - is %v\n", len(texts))
						// break
					}
					content.WriteString(texts[k].S)
					// content.WriteString(" ")
				}
				content.WriteString(" ")
			}

			// target := html.EscapeString(string(respBytes))

			addressablePdf := pdfs[i]
			addressablePdf.Content = content.String()
			numRows, err := gorpx.DBMap().Update(&addressablePdf)
			util.CheckErr(err)
			logx.Printf("%v rows updated", numRows)

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
		FormAction: PathProcessPdfs,

		Gemeinde:   util.EffectiveParam(c, "Gemeinde", "Schwetzingen"),
		Schluessel: util.EffectiveParam(c, "Schluessel", "08 2 26 084"),
		ParamStart: util.EffectiveParam(c, "Start", "1"),
		ParamCount: util.EffectiveParam(c, "Count", "5"),
	}

	err = c.Render("results.html", s)
	util.CheckErr(err)

}
