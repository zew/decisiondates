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

	pdfPage := mdl.PdfJoinPage{}

	//
	//
	srcPageId := util.EffectiveParamInt(c, "srcPageId", -1)
	if srcPageId > 0 {
		sql := `SELECT 
					*
			FROM 			    ` + gorpx.TableName(mdl.Pdf{}) + `  t1
					 INNER JOIN ` + gorpx.TableName(mdl.Page{}) + ` t2 USING(pdf_url)
			WHERE 			1=1
				AND		t2.page_id =  :src_page_id
			`
		args := map[string]interface{}{
			"src_page_id": srcPageId,
		}
		err = gorpx.DBMap().SelectOne(&pdfPage, sql, args)
		util.CheckErr(err)
		display += fmt.Sprintf("For srcPageId %v:  Found %v %v - page %v\n", srcPageId, pdfPage.Pdf.CommunityName,
			pdfPage.Pdf.CommunityKey, pdfPage.Page.Number)
	}

	//
	// Deviating
	dCN := util.EffectiveParam(c, "deviatingCommName", "")
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

		display += fmt.Sprintf("Found %v hit for %v\n", len(communities), dCN)
		for i := 0; i < len(communities); i++ {
			community := communities[i]
			display += fmt.Sprintf("%-5v - %18v: %v\n", community.Id, community.Key, community.Name)
		}
	}

	logx.Printf("---------decision date changed for %q-------", dCN)

	s := struct {
		HTMLTitle string
		Title     string
		Links     []struct{ Title, Url string }

		FormAction     string
		ParamSrcPageId string

		ParamCommName string
		ParamCommKey  string

		ParamDeviatingCommName string

		Url    string
		UrlCmp string

		StructDump template.HTML
		RespBytes  template.HTML
	}{
		HTMLTitle: AppName() + " Enter decision date",
		Title:     AppName() + " Enter decision date",
		Links:     links,

		StructDump: template.HTML(display),

		FormAction: UpdateDecisionDate,

		ParamCommName:          pdfPage.Pdf.CommunityName,
		ParamCommKey:           pdfPage.Pdf.CommunityKey,
		ParamDeviatingCommName: dCN,
		ParamSrcPageId:         util.EffectiveParam(c, "srcPageId", ""),
	}

	err = c.Render("update-decision-date.html", s)
	util.CheckErr(err)

}
