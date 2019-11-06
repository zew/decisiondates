package main

import (
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"

	"github.com/zew/decisiondates/gorpx"
	"github.com/zew/decisiondates/mdl"
	"github.com/zew/logx"
	"github.com/zew/util"
)

func decisionDateEdit(c iris.Context) {

	var err error
	display := ""

	pdfPage := mdl.PdfJoinPage{}

	//
	//
	srcPageId := EffectiveParamInt(c, "SrcPageId", -1)
	srcPdfId := EffectiveParamInt(c, "SrcPdfId", -1)
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
	dCN := EffectiveParam(c, "DeviatingCommName", "")
	dKey, dName, displ2 := deviatingComm(dCN)
	display += displ2
	if dKey != "" {
		pdfPage.Pdf.CommunityKey = dKey
		pdfPage.Pdf.CommunityName = dName
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
			ORDER BY decision_for_year
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
		dec.ForYear = 20
		dec.DecisionDate = "dd.mm.yyyy"
		decisions = append(decisions, dec)
		// dec.ForYear = 2016
		// dec.DecisionDate = "11.11.2015"
		// decisions = append(decisions, dec)

	}

	logx.Printf("---------decision date changed for %q-------", dCN)

	s := struct {
		HTMLTitle string
		Title     string
		FlashMsg  template.HTML
		Links     []struct{ Title, Url string }

		FormAction1 string
		FormAction2 string

		ParamSrcPageId string
		ParamSrcPdfId  string

		ParamCommName string
		ParamCommKey  string

		ParamDeviatingCommName string

		Decisions []mdl.Decision

		StructDump template.HTML
	}{
		HTMLTitle: AppName() + " - Edit decision date",
		Title:     AppName() + " - Edit decision date",
		FlashMsg:  template.HTML(sessions.Get(c).GetFlashString(DecisionDateSave)),
		Links:     links,

		StructDump: template.HTML(display),

		FormAction1: DecisionDateEdit,
		FormAction2: DecisionDateSave,

		ParamCommName: pdfPage.Pdf.CommunityName,
		ParamCommKey:  pdfPage.Pdf.CommunityKey,

		ParamDeviatingCommName: dCN,

		ParamSrcPageId: EffectiveParam(c, "SrcPageId", ""),
		ParamSrcPdfId:  EffectiveParam(c, "SrcPdfId", ""),

		Decisions: decisions,
	}

	err = c.View("decision-date-edit.html", s)
	util.CheckErr(err)

}

func deviatingComm(dCN string) (string, string, string) {

	display := ""

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
		_, err := gorpx.DBMap().Select(&communities, sql, args)
		if err != nil {
			display += fmt.Sprintf("Error selecting deviating communities: %v\n", err)

		}

		if len(communities) > 1 {
			display += fmt.Sprintf("Found %v hit for %v\n", len(communities), dCN)
			for i := 0; i < len(communities); i++ {
				community := communities[i]
				display += fmt.Sprintf("%-18v: %v\n", community.Key, community.Name)
			}
		}
		if len(communities) == 1 {
			display += fmt.Sprintf("Override %-18v: %v\n", communities[0].Key, communities[0].Name)
			return communities[0].Key, communities[0].Name, display
		}
	}

	return "", "", display

}
