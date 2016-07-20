package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kataras/iris"
	"github.com/pbberlin/pdf"

	"github.com/zew/assessmentratedate/gorpx"
	"github.com/zew/assessmentratedate/mdl"
	"github.com/zew/logx"
	"github.com/zew/util"
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
					*
			FROM 			` + gorpx.TableName(mdl.Pdf{}) + ` t1
			WHERE 			1=1
				AND		pdf_id >= :start_id
				AND		pdf_id <  :end_id
				AND		pdf_frequency <= :frequency
			`
		args := map[string]interface{}{
			"start_id":  start,
			"end_id":    end,
			"frequency": maxFrequency,
		}
		gorpx.DBMap().TraceOn("[gorp]", log.New(os.Stdout, "myapp:", log.Lmicroseconds))
		_, err = gorpx.DBMap().Select(&pdfs, sql, args)
		util.CheckErr(err)
		gorpx.DBMap().TraceOff()

		for i := 0; i < len(pdfs); i++ {
			logx.Printf("%-4v  %-4v  %-4v  %v", i, pdfs[i].Id, pdfs[i].Frequency, pdfs[i].Title) // dont print out all fields
		}

		for i := 0; i < len(pdfs); i++ {

			display += fmt.Sprintf("============================\n")
			msg1 := fmt.Sprintf("opening pdf <a href='%v' target='pdf'>id%03v</a> for %v \n", pdfs[i].Url, pdfs[i].Id, pdfs[i].CommunityName)
			display += msg1
			logx.Printf("opening pdf id%03v for %v \n", pdfs[i].Id, pdfs[i].CommunityName)

			if pdfs[i].Frequency > maxFrequency {
				msg2 := fmt.Sprintf("  skipping due to frequency %v \n\n", pdfs[i].Frequency)
				display += msg2
				logx.Printf(msg2)
				continue
			}

			req, err := http.NewRequest("GET", pdfs[i].Url, nil)
			util.CheckErr(err)

			strUrl = pdfs[i].Url

			resp, err := util.HttpClient().Do(req)
			if err != nil {
				logx.Printf("client.Do(): %v", err)
				continue
			}
			defer resp.Body.Close()

			respBytes, err = ioutil.ReadAll(resp.Body)
			util.CheckErr(err)

			// ioutil.WriteFile(fmt.Sprintf("pdfNumber%03v.pdf", pdfs[i].Id), respBytes, os.FileMode(777))

			rdr := bytes.NewReader(respBytes)
			rdr2, err := pdf.NewReader(rdr, int64(len(respBytes)))
			if err != nil {
				logx.Printf("%v", err)
				continue
			}

			numPages, err := func() (numP int, err error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("calling numpages recover: %v", r)
					}
				}()
				numP = rdr2.NumPage()
				return
			}()
			if err != nil {
				logx.Printf("%v", err)
				continue
			}

			logx.Printf(" found %v pages\n", numPages)
			for j := 1; j <= numPages; j++ {
				if j >= maxPages {
					logx.Printf("  not opening more than %v pages", maxPages)
					break
				}

				// logx.Printf("   opening page %v\n", j)

				page := rdr2.Page(j)
				cn, err := extractContent(&page)
				if err != nil {
					logx.Printf("Page_%002v: %v", j, err)
					continue
				}
				texts := cn.Text
				cnBf := bytes.Buffer{} // content buffer
				for k := 0; k < len(texts); k++ {
					cnBf.WriteString(texts[k].S)
				}

				p := mdl.Page{}
				p.Url = pdfs[i].Url
				p.Number = j
				p.Content = cnBf.String()

				if false {
					p.Content = strings.TrimSpace(p.Content)
					p.Content = strings.Join(strings.Fields(p.Content), " ") // strip all white space
				}
				err = gorpx.DBMap().Insert(&p)
				if err != nil {
					errStr := fmt.Sprintf("%v", err)
					if !strings.Contains(errStr, "Error 1062: Duplicate entry") {
						logx.Printf("insert error %v; trying updated", err)
					}

					args := map[string]interface{}{
						"page_url":    p.Url,
						"page_number": p.Number,
					}

					primKey, err := gorpx.DBMap().SelectInt("select page_id from "+
						gorpx.TableName(p)+
						" where page_url = :page_url AND page_number = :page_number",
						args)

					util.CheckErr(err)
					if primKey > 0 {
						p.Id = int(primKey)
						numRows, err := gorpx.DBMap().Update(&p)
						util.CheckErr(err)
						if numRows > 0 {
							logx.Printf("     %v rows updated - page id %v", numRows, p.Id)
						}
					}

				}

			}

		}
		logx.Printf("---------text extraction finished for--%v-%v-----", start, end)

	}

	s := struct {
		HTMLTitle string
		Title     string
		Links     []struct{ Title, Url string }

		FormAction string
		ParamStart string
		ParamCount string

		Url    string
		UrlCmp string

		StructDump template.HTML
		RespBytes  template.HTML
	}{
		HTMLTitle: AppName() + " - download pdfs and extract their text content",
		Title:     AppName() + " - download pdfs and extract their text content",
		Links:     links,

		StructDump: template.HTML(display),
		// RespBytes:  template.HTML(string(respBytes)),

		Url:        strUrl,
		FormAction: PathProcessPdfs,

		ParamStart: util.EffectiveParam(c, "Start", "0"),
		ParamCount: util.EffectiveParam(c, "Count", "3"),
	}

	err = c.Render("results.html", s)
	util.CheckErr(err)

}

func extractContent(p *pdf.Page) (cnt *pdf.Content, err error) {
	defer func() {
		if r := recover(); r != nil {
			rs := fmt.Sprintf("%v", r)
			rs = util.EnsureUtf8(rs)
			rs = strings.TrimSpace(rs)
			rs = util.Ellipsoider(rs, 200)
			if rs == "malformed PDF: reading at offset 0: stream not present" {
				err = fmt.Errorf("extractContent() recover: no stream at offset 0")
			} else if strings.HasPrefix(rs, "malformed hex string") {
				err = fmt.Errorf("extractContent() recover: malformed hex string")
			} else {
				err = fmt.Errorf("extractContent() recover: %v", r)
			}
		}
	}()
	cntDeref := p.Content()
	cnt = &cntDeref
	return
}
