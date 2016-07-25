package main

import (
	"github.com/kataras/iris"
	"github.com/zew/util"
)

func index(c *iris.Context) {

	var err error
	s := struct {
		HTMLTitle string
		Title     string
		Links     []struct{ Title, Url string }
	}{
		HTMLTitle: AppName() + " - Main",
		Title:     AppName() + " - Main",
		Links:     links,
	}

	err = c.Render("index.html", s)
	util.CheckErr(err)

}
