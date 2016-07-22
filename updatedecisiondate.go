package main

import (
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/kataras/iris"

	"github.com/zew/assessmentratedate/gorpx"
	"github.com/zew/assessmentratedate/mdl"
	"github.com/zew/logx"
	"github.com/zew/util"
)

func updateDecisionDate(c *iris.Context) {

	var err error
	display := ""

	{
		type FormDecisions struct {
			CommName          string
			CommKey           string
			SrcPageId         int
			SrcPdfId          int
			DeviatingCommName string
			Submit1, Submit2  string
			Decisions         []mdl.Decision
		}
		frm := FormDecisions{}
		if true || c.IsPost() {
			err = c.ReadForm(&frm)
			if err != nil {
				logx.Println("error parsing form", err.Error())
			}
			logx.Printf("form was: %v %v", frm.CommKey, frm.CommName)
			for idx, dec := range frm.Decisions {
				logx.Printf("%v  %v %v %v %v", idx, dec.Id, dec.ForYear, dec.DecisionDate, dec.PageId)
			}
		}
	}

	pdfPage := mdl.PdfJoinPage{}

	//
	//
	srcPageId := util.EffectiveParamInt(c, "SrcPageId", -1)
	srcPdfId := util.EffectiveParamInt(c, "SrcPdfId", -1)
	if srcPageId > 0 {
		sql := `SELECT 
					*
			FROM 			    ` + gorpx.TableName(mdl.Pdf{}) + `  t1
					 INNER JOIN ` + gorpx.TableName(mdl.Page{}) + ` t2 USING(pdf_url)
			WHERE 			1=1
				AND		t1.pdf_id  =  :src_pdf_id
				AND		t2.page_id =  :src_page_id
			`
		args := map[string]interface{}{
			"src_pdf_id":  srcPdfId,
			"src_page_id": srcPageId,
		}
		err = gorpx.DBMap().SelectOne(&pdfPage, sql, args)
		util.CheckErr(err)
		display += fmt.Sprintf("%-18v: %v    page %v    (srcPageId %v) \n",
			pdfPage.Pdf.CommunityKey, pdfPage.Pdf.CommunityName,
			pdfPage.Page.Number, srcPageId)
	}

	//
	// Deviating
	dCN := util.EffectiveParam(c, "DeviatingCommName", "")
	if dCN != "" {
		dCN = strings.TrimSpace(dCN)
		dCN = strings.ToLower(dCN)
		communities := []mdl.Community{}
		sql := `SELECT 
						*
				FROM 			` + gorpx.TableName(mdl.Community{}) + ` t1
				WHERE 			1=1
					AND		(  LOWER(community_name) = :community_name OR 1 )
					AND		LOWER(community_name) like :community_name_wildcarded
				`
		communityNameWildcarded := fmt.Sprintf("%%%v%%", dCN) // this is not working :(
		communityNameWildcarded = "%%" + dCN + "%%"           // this works
		log.Print("communityNameWildcarded", "\t", dCN, "\t", communityNameWildcarded)
		args := map[string]interface{}{
			"community_name":            dCN,
			"community_name_wildcarded": communityNameWildcarded,
		}
		_, err = gorpx.DBMap().Select(&communities, sql, args)
		util.CheckErr(err)

		if len(communities) > 1 {
			display += fmt.Sprintf("Found %v hit for %v\n", len(communities), dCN)
			for i := 0; i < len(communities); i++ {
				community := communities[i]
				display += fmt.Sprintf("%-18v: %v\n", community.Key, community.Name)
			}
		}
		if len(communities) == 1 {
			display += fmt.Sprintf("Override %v\n", pdfPage.Pdf.CommunityName)
			display += fmt.Sprintf("%-18v: %v\n", communities[0].Key, communities[0].Name)
			pdfPage.Pdf.CommunityName = communities[0].Name
			pdfPage.Pdf.CommunityKey = communities[0].Key
		}
	}

	//
	//
	decisions := []mdl.Decision{}
	if pdfPage.Pdf.CommunityKey != "" {
		sql := `SELECT 
					*
			FROM 			    ` + gorpx.TableName(mdl.Decision{}) + `  t1
			WHERE 			1=1
				AND		t1.community_key =  :community_key
			`
		args := map[string]interface{}{
			"community_key": pdfPage.Pdf.CommunityKey,
		}
		_, err := gorpx.DBMap().Select(&decisions, sql, args)
		util.CheckErr(err)
		for _, decision := range decisions {
			display += fmt.Sprintf("  decision: %v - %v - derived from %v\n", decision.ForYear, decision.DecisionDate, decision.PageId)
		}

		//
		// some empty
		dec := mdl.Decision{}
		dec.PageId = srcPageId
		dec.ForYear = 2015
		dec.DecisionDate = "01.01.2015"
		decisions = append(decisions, dec)
		dec.ForYear = 2016
		dec.DecisionDate = "11.11.2015"
		decisions = append(decisions, dec)

	}

	logx.Printf("---------decision date changed for %q-------", dCN)

	s := struct {
		HTMLTitle string
		Title     string
		Links     []struct{ Title, Url string }

		FormAction string

		ParamSrcPageId string
		ParamSrcPdfId  string

		ParamCommName string
		ParamCommKey  string

		ParamDeviatingCommName string

		Decisions []mdl.Decision

		StructDump template.HTML
	}{
		HTMLTitle: AppName() + " Enter decision date",
		Title:     AppName() + " Enter decision date",
		Links:     links,

		StructDump: template.HTML(display),

		FormAction: UpdateDecisionDate,

		ParamCommName: pdfPage.Pdf.CommunityName,
		ParamCommKey:  pdfPage.Pdf.CommunityKey,

		ParamDeviatingCommName: dCN,

		ParamSrcPageId: util.EffectiveParam(c, "SrcPageId", ""),
		ParamSrcPdfId:  util.EffectiveParam(c, "SrcPdfId", ""),

		Decisions: decisions,
	}

	err = c.Render("update-decision-date.html", s)
	util.CheckErr(err)

}
