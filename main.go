package main

import (
	_ "pyg/routers"
	"github.com/astaxie/beego"
	"pyg/models"
)

func main() {
	models.Init()
	beego.Run()
}

