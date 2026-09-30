package controllers

type HomeController struct {
	BaseController
}

func (c *HomeController) Get() {
	c.Layout = "layout.tpl"
	c.Data["Title"] = "Find events"
	c.Data["Script"] = "/static/js/home.js"
	c.TplName = "home.tpl"
}
