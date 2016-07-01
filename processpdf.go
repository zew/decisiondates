package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/kataras/iris"
	"github.com/pbberlin/pdf"

	"github.com/zew/assessmentratedate/gorpx"
	"github.com/zew/assessmentratedate/logx"
	"github.com/zew/assessmentratedate/mdl"
	"github.com/zew/assessmentratedate/util"
)

const maxPages = 217

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
		gorpx.DBMap().TraceOn("[gorp]", log.New(os.Stdout, "myapp:", log.Lmicroseconds))
		_, err = gorpx.DBMap().Select(&pdfs, sql, args)
		util.CheckErr(err)
		gorpx.DBMap().TraceOff()

		for i := 0; i < len(pdfs); i++ {
			logx.Printf("%-4v  %55v %v", i, util.UpTo(pdfs[i].Url, 70), pdfs[i].Title) // dont print out all fields
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

			ioutil.WriteFile(fmt.Sprintf("pdfNumber%03v.pdf", pdfs[i].Id), respBytes, os.FileMode(777))

			rdr := bytes.NewReader(respBytes)

			msg1 := fmt.Sprintf("opening pdf #%-2v %v\n", i, util.UpToR(pdfs[i].Url, 70))
			logx.Printf(msg1)
			display += msg1

			rdr2, err := pdf.NewReader(rdr, int64(len(respBytes)))
			if err != nil {
				logx.Printf("%v", err)
				continue
			}

			content := bytes.Buffer{}

			numPages := rdr2.NumPage()
			logx.Printf(" found %v pages\n", numPages)
			for j := 1; j <= numPages; j++ {
				if j >= maxPages {
					logx.Printf("  not opening more than %v pages", maxPages)
					break
				}
				content.WriteString(fmt.Sprintf("\npdf_page_%002v\n", j))
				// logx.Printf("   opening page %v\n", j)
				page := rdr2.Page(j)

				cn, err := extractContent(&page)
				if err != nil {
					logx.Printf("%v", err)
					continue
				}
				if cn == nil {
					logx.Printf("page content is nil")
					continue
				}
				texts := cn.Text
				for k := 0; k < len(texts); k++ {
					content.WriteString(texts[k].S)
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
		// RespBytes:  template.HTML(string(respBytes)),

		Url:        strUrl,
		FormAction: PathProcessPdfs,

		Gemeinde:   util.EffectiveParam(c, "Gemeinde", "Schwetzingen"),
		Schluessel: util.EffectiveParam(c, "Schluessel", "08 2 26 084"),
		ParamStart: util.EffectiveParam(c, "Start", "0"),
		ParamCount: util.EffectiveParam(c, "Count", "3"),
	}

	err = c.Render("results.html", s)
	util.CheckErr(err)

}

func extractContent(p *pdf.Page) (cnt *pdf.Content, err error) {
	defer func() {
		if r := recover(); r != nil {
			// panic: malformed PDF: reading at offset 0: stream not present
			err = fmt.Errorf("Recovered while reading page content: %v", r)
			// logx.Printf("in defer: %v", err)
		}
	}()
	cntDeref := p.Content()
	cnt = &cntDeref
	return
}
