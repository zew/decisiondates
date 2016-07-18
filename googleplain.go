package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kataras/iris"
	"github.com/zew/assessmentratedate/config"
	"github.com/zew/assessmentratedate/gorpx"
	"github.com/zew/assessmentratedate/logx"
	"github.com/zew/assessmentratedate/mdl"
	"github.com/zew/assessmentratedate/util"
)

// Here we get the plain JSON response.
// The request needs to be decorated with a searchEngineId and a
// and with some app engine credentials
//
// There is no need for a special oauth2 client.
func plainJsonResponse(c *iris.Context) (string, []byte, error) {

	display := ""
	respBytes := []byte{}
	community := mdl.Community{}
	strUrl := ""

	myUrl := url.URL{}
	myUrl.Host = "www.googleapis.com"
	myUrl.Path = "customsearch/v1"
	myUrl.Scheme = "https"
	logx.Printf("host is %v", myUrl.String())

	// https://developers.google.com/apis-explorer/#p/customsearch/v1/search.cse.list?q=Schwetzingen&_h=1&

	vals := map[string]string{
		"key":   config.Config.AppEngineServerKey,
		"cx":    config.Config.GoogleCustomSearchId,
		"q":     util.EffectiveParam(c, "Gemeinde", "Villingen-Schwenningen"),
		"start": util.EffectiveParam(c, "Start", "1"),
		"num":   util.EffectiveParam(c, "Count", "5"),
		"safe":  "off",
	}

	queryStr := ""
	for k, v := range vals {
		queryStr += fmt.Sprintf("%v=%v&", k, v)
	}
	logx.Printf("queryStr is %v", queryStr)

	strUrl = myUrl.String() + "/?" + queryStr
	req, err := http.NewRequest("GET", strUrl, nil)
	util.CheckErr(err)

	resp, err := util.HttpClient().Do(req)
	util.CheckErr(err)
	defer resp.Body.Close()

	respBytes, err = ioutil.ReadAll(resp.Body)
	util.CheckErr(err)

	// Parse
	if err != nil {
		c.Text(200, err.Error())
		return display, respBytes, err
	}

	err = gorpx.DBMap().Insert(&community)
	if err != nil {
		c.Text(200, err.Error())
	}

	display = util.IndentedDump(community)
	// c.Text(200, display)

	return display, respBytes, nil
}
