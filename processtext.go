package main

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"

	"github.com/kataras/iris"

	"github.com/zew/assessmentratedate/gorpx"
	"github.com/zew/assessmentratedate/logx"
	"github.com/zew/assessmentratedate/mdl"
	"github.com/zew/assessmentratedate/util"
)

func processText(c *iris.Context) {

	var err error
	display := ""
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
			`
		args := map[string]interface{}{
			"start_id": start,
			"end_id":   end,
		}
		_, err = gorpx.DBMap().Select(&pdfs, sql, args)
		util.CheckErr(err)

		r1, err := regexp.Compile("Hebes([aä]+)tz[e]")
		util.CheckErr(err)

		// original regex: ("[0-9]{2}[./ ]+[0-9]{2}[./ ]+[0-9]{4}")
		weekdays := "1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|25|26|27|28|29|30|31|01|02|03|04|05|06|07|08|09"
		monthsLong := "Januar|Februar|März|April|Mai|Juni|Juli|August|September|Oktober|November|Dezember"
		monthsShort := "Jan|Feb|Mrz|Apr|Mai|Jun|Jul|Aug|Sept|Sep|Okt|Nov|Dez"
		monthsNumbered := "1|2|3|4|5|6|7|8|9|01|02|03|04|05|06|07|08|09|10|11|12"
		yearsLong := "2012|2013|2014|2015|2016"
		all := fmt.Sprintf("((%v)[./\\s]+(%v|%v|%v)[./\\s]+(%v))[^0-9]+", weekdays, monthsLong, monthsShort, monthsNumbered, yearsLong)
		r2, err := regexp.Compile(all)
		util.CheckErr(err)

		for i := 0; i < len(pdfs); i++ {

			// r1.MatchString("peach")
			addressablePdf := pdfs[i]
			display += fmt.Sprintf("<a href='%v' target='pdf'>%v: %v</a>\n", addressablePdf.Url, addressablePdf.CommunityName, addressablePdf.Title)

			if addressablePdf.Frequency > 2 {
				display += fmt.Sprintf("skipping due to frequency %v \n\n", addressablePdf.Frequency)
				continue
			}

			{
				matchPos := r1.FindAllStringIndex(addressablePdf.Content, -1)
				addressablePdf.Snippet1 = fmt.Sprintf("%v", matchPos)
				addressablePdf.Snippet2 = ""
				for idx, occurrence := range matchPos {
					sn := snippetIt(occurrence, addressablePdf.Content, 10, 30)
					addressablePdf.Snippet2 += fmt.Sprintf("#%02v: Pos:%v  \n%v  \n\n", idx, occurrence, sn)
				}
				// display += addressablePdf.Snippet2
			}

			// search for the date
			{
				matchPosAll := r2.FindAllStringSubmatchIndex(addressablePdf.Content, -1)
				addressablePdf.Snippet3 = ""
				for idx, occurrence := range matchPosAll {
					sn := snippetIt(occurrence[2:4], addressablePdf.Content, 6, 12)
					addressablePdf.Snippet3 += fmt.Sprintf("#%03v: %v  %v  \n\n", idx, formatPos(occurrence, len(addressablePdf.Content)), sn)

				}
				display += addressablePdf.Snippet3
			}

			//
			//
			numRows, err := gorpx.DBMap().Update(&addressablePdf)
			if err != nil {
				display += fmt.Sprintf("Error during update: %v \n%v\n%v", err, &addressablePdf.Snippet2, &addressablePdf.Snippet3)
			}
			// util.CheckErr(err)
			logx.Printf("%v rows updated", numRows)
			display += "\n\n"

		}
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
		HTMLTitle: AppName() + " find text passages",
		Title:     AppName() + " find text passages",
		Links:     links,

		StructDump: template.HTML(display),
		// RespBytes:  template.HTML(string(respBytes)),

		Url:        strUrl,
		FormAction: PathProcessText,

		Gemeinde:   util.EffectiveParam(c, "Gemeinde", "Schwetzingen"),
		Schluessel: util.EffectiveParam(c, "Schluessel", "08 2 26 084"),
		ParamStart: util.EffectiveParam(c, "Start", "0"),
		ParamCount: util.EffectiveParam(c, "Count", "3"),
	}

	err = c.Render("results.html", s)
	util.CheckErr(err)

}

func formatPos(occurrence []int, fullLen int) string {

	pct := float64(occurrence[0]) / float64(fullLen) * 100
	return fmt.Sprintf("%02f%%", pct)
	return fmt.Sprintf("%04.1f%%", pct)

}

func snippetIt(occurrence []int, haystack string, before int, after int) string {

	l := len(haystack)
	_ = l
	start := occurrence[0]
	stop := occurrence[1]

	start -= before
	if start < 0 {
		start = 0
	}

	stop += after
	if stop > l {
		stop = l
	}

	ret := bytes.Buffer{}
	// looping over possibly invalid utf-8 sequences
	// "... If the iteration encounters an invalid UTF-8 sequence, the second value will be 0xFFFD, ...""
	cnt := 0
	max := before + occurrence[1] - occurrence[0] + after
	for idx, codepoint := range haystack[start:stop] {

		if idx == before {
			// ret.WriteRune(rune(32)) // enclose into extra spaces
			ret.WriteString("<b>")
		}
		if idx == before+occurrence[1]-occurrence[0] {
			ret.WriteString("</b> ")
		}

		ret.WriteRune(codepoint)

		cnt++
		if cnt > max {
			break
		}
	}

	return ret.String()
	return haystack[start:stop]

}
