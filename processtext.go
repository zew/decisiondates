package main

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"

	"github.com/kataras/iris"

	"github.com/zew/assessmentratedate/gorpx"
	"github.com/zew/assessmentratedate/mdl"
	"github.com/zew/assessmentratedate/util"
	"github.com/zew/awis/logx"
)

func processText(c *iris.Context) {

	var err error
	display := ""
	strUrl := ""

	r1, err := regexp.Compile("Hebes([aä]+)tz[e]")
	util.CheckErr(err)

	str1 := "amtliche Bekanntmachung|Amtsblatt|Anzeiger|Bürgerhaushalt"
	str2 := "Gewerbesteuer|Gemeindeanzeiger|Gemeindeblatt|Gemeinderatsbeschluß|Grundsteuer"
	str3 := "Haushaltrede|Haushaltsdokument|Haushaltsplan|Haushaltssanierungsplan|Haushaltssatzung|Hebesatzsatzung"
	str4 := "Jahresabschluss|Jahresabschluß|Mitteilungsblatt|Nachhaltigkeitssatzung"
	str5 := "Protokoll|Sitzung|Stadtanzeiger"
	all2 := fmt.Sprintf("(?i)%v|%v|%v|%v|%v", str1, str2, str3, str4, str5)
	r2, err := regexp.Compile(all2)
	util.CheckErr(err)

	//
	// original regex: ("[0-9]{2}[./ ]+[0-9]{2}[./ ]+[0-9]{4}")
	weekdays := "1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|25|26|27|28|29|30|31|01|02|03|04|05|06|07|08|09"
	monthsLong := "Januar|Februar|März|April|Mai|Juni|Juli|August|September|Oktober|November|Dezember"
	monthsShort := "Jan|Feb|Mrz|Apr|Mai|Jun|Jul|Aug|Sept|Sep|Okt|Nov|Dez"
	monthsNumbered := "1|2|3|4|5|6|7|8|9|01|02|03|04|05|06|07|08|09|10|11|12"
	yearsLong := "2010|2011|2012|2013|2014|2015|2016"
	all3 := fmt.Sprintf("((%v)[./\\s]+(%v|%v|%v)[./\\s]+(%v))[^0-9]+", weekdays, monthsLong, monthsShort, monthsNumbered, yearsLong)
	r3, err := regexp.Compile(all3)
	util.CheckErr(err)

	//
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
		_, err = gorpx.DBMap().Select(&pdfs, sql, args)
		util.CheckErr(err)

		for i := 0; i < len(pdfs); i++ {

			// r1.MatchString("peach")
			addressablePdf := pdfs[i]

			if addressablePdf.Frequency > maxFrequency {
				// display += fmt.Sprintf(
				// 	"Skipping <a href='%v' target='pdf' >%v: %v</a> &nbsp; due to frequency %v \n ",
				// 	addressablePdf.Url, addressablePdf.CommunityName, addressablePdf.Title, addressablePdf.Frequency)
				continue
			}

			var fndPdf1, fndPdf2, fndPdf3 bool
			var pdfDisp string

			pages := []mdl.Page{}
			sql := `SELECT 	*
			FROM 			` + gorpx.TableName(mdl.Page{}) + ` t1
			WHERE 			1=1
				AND		page_url = :page_url   `
			args := map[string]interface{}{
				"page_url": addressablePdf.Url,
			}
			_, err = gorpx.DBMap().Select(&pages, sql, args)
			util.CheckErr(err)

			addressablePdf.Snippet1 = ""
			addressablePdf.Snippet2 = ""
			addressablePdf.Snippet3 = ""

			for j := 0; j < len(pages); j++ {

				p := pages[j]
				var fndOnPage1, fndOnPage2, fndOnPage3 bool

				// search for the hebesatz
				var sn1p string
				{
					matchPos := r1.FindAllStringIndex(p.Content, -1)
					addressablePdf.Snippet1 = fmt.Sprintf("%v", matchPos)
					if len(matchPos) > 2 {
						matchPos = append(matchPos[:1], matchPos[len(matchPos)-1:]...)
					}
					for idx, occurrence := range matchPos {
						fndOnPage1 = true
						fndPdf1 = true
						sn := snippetIt(occurrence, p.Content, 20, 110)
						sn1p += fmt.Sprintf("#%03v: %02v%%  %v  \n\n", idx, 100*occurrence[0]/len(p.Content), sn)
					}
					addressablePdf.Snippet2 += sn1p
				}

				// search for "amtliche Mitteilung, etc"
				var sn2p string
				{
					matchPos := r2.FindAllStringIndex(p.Content, -1)
					if len(matchPos) > 2 {
						matchPos = append(matchPos[:1], matchPos[len(matchPos)-1:]...)
					}
					for idx, occurrence := range matchPos {
						// fndOnPage2 = true
						fndPdf2 = true
						sn := snippetIt(occurrence, p.Content, 20, 110)
						_ = idx
						_ = sn
						// sn2p += fmt.Sprintf("#%03v: %02v%%  %v  \n\n", idx, 100*occurrence[0]/len(p.Content), sn)
					}
					addressablePdf.Snippet2 += sn2p
				}

				// search for the date
				var sn3p string
				{
					matchPosAll := r3.FindAllStringSubmatchIndex(p.Content, -1)
					if len(matchPosAll) > showMaxXDates {
						matchPosAll = append(matchPosAll[:showMaxXDates/2], matchPosAll[len(matchPosAll)-showMaxXDates/2:]...)
					}
					for idx, occurrence := range matchPosAll {
						fndOnPage3 = true
						fndPdf3 = true
						sn := snippetIt(occurrence[2:4], p.Content, 12, 40) // the second sub-match
						sn3p += fmt.Sprintf("#%03v: %v  %v    \n\n", idx, formatPos(occurrence, len(p.Content)), sn)
					}
					addressablePdf.Snippet3 += sn3p
				}

				if fndOnPage1 || fndOnPage2 || fndOnPage3 {
					pdfDisp += fmt.Sprintf("<a href='%v#page=%v' target='pdf'>Seite %02v</a>\n",
						addressablePdf.Url, p.Number, p.Number)
					pdfDisp += util.Ellipsoider(sn1p, 1800)
					pdfDisp += util.Ellipsoider(sn2p, 1800)
					pdfDisp += util.Ellipsoider(sn3p, 1800)
					pdfDisp += "\n"
				}

			}

			//
			//
			numRows, err := gorpx.DBMap().Update(&addressablePdf)
			if err != nil {
				display += fmt.Sprintf("Error during update: %v \n%v\n%v", err, &addressablePdf.Snippet2, &addressablePdf.Snippet3)
				continue
			}
			if numRows > 0 {
				logx.Printf("%v rows updated; pdf_id %-5v", numRows, addressablePdf.Id)

			}

			if fndPdf1 && fndPdf2 && fndPdf3 {
				display += fmt.Sprintf("<a href='%v' target='pdf'>%v: %v</a>\n",
					addressablePdf.Url, addressablePdf.CommunityName, addressablePdf.Title)
				display += pdfDisp + "<hr/>\n\n"
			}

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
	return fmt.Sprintf("%02.0f%% ", pct)
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
	ret.WriteString("</b> ")

	return ret.String()

}
