package controllers

import (
	"github.com/astaxie/beego"
	"pyg/models"
	"math"
	"fmt"
)

type GoodsController struct {
	beego.Controller
}

//展示主界面
func (this *GoodsController) ShowIndex() {

	//获取用户名
	userName := this.GetSession("userName")
	if userName != nil {
		this.Data["userName"] = userName.(string)
	} else {
		this.Data["userName"] = ""
	}
	goodsTypes := models.GetAllMenu()
	this.Data["goodsTypes"] = goodsTypes
	this.TplName = "index.html"
}

//展示生鲜界面
func (this *GoodsController) ShowIndexSx() {
	//获取所有类型
	types, err1 := models.GetAllType()
	//获取首页轮播图
	banners, err2 := models.GetAllBanner()
	//获取首页促销商品
	promotinos, err3 := models.GetAllPromotion()
	if err1 != nil || err2 != nil || err3 != nil {
		this.Data["errmsg"] = "获取信息失败"
		this.TplName = "index_sx.html"
		return
	}
	goods := models.GetAllIndexGoods(types)
	this.Data["goods"] = goods
	this.Data["types"] = types
	this.Data["banners"] = banners
	this.Data["promotions"] = promotinos
	this.TplName = "index_sx.html"
}

//获取详情
func (this *GoodsController) ShowDetail() {
	id, err := this.GetInt("id")
	if err != nil {
		this.Data["errmsg"] = "链接错误"
		this.TplName = "index_sx.html"
		return
	}
	//得到页面展示的商品种类详情
	err, goodsSKU := models.GetGoodsDetail(id)
	if err != nil {
		this.Data["errmsg"] = "查询数据库错误"
		this.TplName = "index_sx.html"
		return
	}
	//获取新品
	newGoods, err := models.GetNewGoods(goodsSKU.GoodsType.Name)
	if err != nil {
		this.Data["errmsg"] = "新品推荐错误"
		this.TplName = "index_sx.html"
		return
	}
	userName := this.GetSession("userName")
	if userName != nil {
		err := models.SaveHistory(userName.(string), id)
		if err != nil {
			this.Data["errmsg"] = "历史浏览记录存储失败"
			this.TplName = "index_sx.html"
			return
		}
	}
	this.Data["newGoods"] = newGoods
	this.Data["goodsSKU"] = goodsSKU
	this.TplName = "detail.html"
}

func ShowPageIndex(pageCount, pageIndex int) []int {
	var pages []int

	//如果总页码小于等于5页,有多少页展示多少页
	if pageCount <= 5 {
		for i := 1; i <= pageCount; i++ {
			pages = append(pages, i)
		}
		return pages
	}
	//如果页码大于5页,当前页是前三页
	if pageIndex <= 3 {
		for i := 1; i <= 5; i++ {
			pages = append(pages, i)
		}
		return pages
	}
	//如果页码大于5页,当前页是后三页
	if pageIndex >= pageCount-2 {
		for i := pageCount - 4; i <= pageCount; i++ {
			pages = append(pages, i)
		}
		return pages
	}
	//如果页码大于5页是中间页
	for i := pageIndex - 2; i <= pageIndex+2; i++ {
		pages = append(pages, i)
	}
	return pages
}

//获取展示列表页
func (this *GoodsController) ShowList() {
	typeName := this.GetString("typeName")
	if typeName == "" {
		this.Data["errmsg"] = "链接错误"
		this.TplName = "index_sx.html"
		return
	}
	types, err1 := models.GetAllType()
	if err1 != nil {
		this.Data["errmsg"] = "获取商品种类失败"
		this.TplName = "index_sx.html"
		return
	}
	sort := this.GetString("sort")
	//AllGoodsTypes, err := models.GetTypeGoods(typeName, sort)
	//if err != nil {
	//	this.Data["errmsg"] = "查询错误"
	//	this.TplName = "index_sx.html"
	//	return
	//}
	newGoods, err := models.GetNewGoods(typeName)
	if err != nil {
		this.Data["errmsg"] = "查询新品错误"
		this.TplName = "index_sx.html"
		return
	}
	//获取总页码
	count, err := models.GetGoodsCount(typeName)

	pageSize := 1

	pageCount := math.Ceil(float64(count) / float64(pageSize))

	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	pages := ShowPageIndex(int(pageCount), pageIndex)
	//获取对应页码的数据

	goodsLimit, err := models.GetPageIndexGoods(pageSize, pageIndex, typeName, sort)
	if err != nil {
		this.Data["errmsg"] = "获取分页错误"
		this.TplName = "list.html"
		return
	}
	//获取上一页
	prePage := pageIndex - 1
	if pageIndex <= 1 {
		prePage = 1
	}
	//获取下一页
	nextPage := pageIndex + 1
	if pageIndex >= int(pageCount) {
		nextPage = int(pageCount)
	}

	this.Data["prePage"] = prePage
	this.Data["nextPage"] = nextPage
	this.Data["pageIndex"] = pageIndex
	this.Data["pages"] = pages
	this.Data["types"] = types
	this.Data["sort"] = sort
	this.Data["newGoods"] = newGoods
	this.Data["AllGoodsTypes"] = goodsLimit
	this.Data["typeName"] = typeName
	this.TplName = "list.html"
}

func (this *GoodsController) SearchGoods() {
	goodsName := this.GetString("goodsName")
	var goods []models.GoodsSKU
	var err error
	if goodsName == "" {
		goods, err = models.GetAllGoods()
	} else {
		goods, err = models.SearchGoods(goodsName)
	}
	if err!=nil{
		fmt.Println("errmsg:=",err)
	}
	this.Data["goods"]=goods
	this.TplName="search.html"

}

