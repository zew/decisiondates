package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kataras/iris"
	"github.com/zew/assessmentratedate/gorpx"
	"github.com/zew/assessmentratedate/mdl"
	"github.com/zew/util"
)

func decisionDateSave(c *iris.Context) {

	var msg bytes.Buffer

	defer func() {
		c.SetFlash(DecisionDateSave, msg.String())
		c.Redirect(
			Pref(DecisionDateEdit) +
				fmt.Sprintf(
					"?SrcPageId=%v&SrcPdfId=%v&DeviatingCommName=%v",
					util.EffectiveParam(c, "SrcPageId"),
					util.EffectiveParam(c, "SrcPdfId"),
					util.EffectiveParam(c, "DeviatingCommName"),
				),
		)
	}()

	type FormDecisions struct {
		CommName          string
		CommKey           string
		SrcPageId         int
		SrcPdfId          int
		DeviatingCommName string
		Submit2           string
		Decisions         []mdl.Decision
	}

	frm := FormDecisions{}
	if c.IsPost() {
		err := c.ReadForm(&frm)
		if err != nil {
			msg.WriteString(fmt.Sprintf("Each form field must be accomodated in the struct! %v\n", err))
			return
		}

		msg.WriteString(fmt.Sprintf("Form was: %q - %q\n", frm.CommKey, frm.CommName))

		for idx, dec := range frm.Decisions {

			dec.CommunityKey = frm.CommKey
			dec.CommunityName = frm.CommName

			if dec.ForYear < 22 || dec.DecisionDate == "dd.mm.yyyy" {
				continue
			}

			if dec.DecisionDate == "deleteme" && dec.Id > 0 {
				numRows, err := gorpx.DBMap().Delete(&dec)
				if err != nil {
					msg.WriteString(fmt.Sprintf("Delete error. %v\n", err))
				}
				if numRows > 0 {
					msg.WriteString(fmt.Sprintf("%v rows deleted - page id %v\n", numRows, &dec.Id))
				}
				continue
			}

			msg.WriteString(
				fmt.Sprintf("#%v id%4v %6v %v %v \n", idx, dec.Id, dec.ForYear, dec.DecisionDate, dec.PageId),
			)

			if dec.Id < 1 {
				err = gorpx.DBMap().Insert(&dec)
				if err != nil {
					if !strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
						msg.WriteString(fmt.Sprintf("insert error was %v; \n", err))
						continue
					} else {
						pk := primKey(&dec, dec.CommunityKey, dec.ForYear)
						if pk > 0 {
							dec.Id = int(pk)
							buf1 := updateDec(&dec)
							msg.Write(buf1.Bytes())
						}
					}
				}
			} else {
				buf1 := updateDec(&dec)
				msg.Write(buf1.Bytes())
			}
			//
		}
	}

}

func primKey(pdec *mdl.Decision, commKey string, forYear int) int64 {

	gorpx.TraceOn()
	primKey, err := gorpx.DBMap().SelectInt("select decision_id from "+
		gorpx.TableName(mdl.Decision{})+
		` where 
			community_key = :community_key 
		AND decision_for_year = :decision_for_year`,
		map[string]interface{}{
			"community_key":     commKey,
			"decision_for_year": forYear,
		},
	)
	gorpx.TraceOff()
	util.CheckErr(err)
	return primKey

}

func updateDec(pdec *mdl.Decision) bytes.Buffer {
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("trying update for id %v; \n", pdec.Id))
	numRows, err := gorpx.DBMap().Update(pdec)
	if err != nil {
		msg.WriteString(fmt.Sprintf("Update error. %v\n", err))
	}
	if numRows > 0 {
		msg.WriteString(fmt.Sprintf("%v rows updated - page id %v\n", numRows, pdec.Id))
	}
	return msg
}
