package controllers

import (
	"github.com/astaxie/beego"
	"pyg/models"
	"fmt"
)

type CartController struct {
	beego.Controller
}

//添加购物车数量
func (this *CartController) AddCart() {
	userName := this.GetSession("userName")
	//定义容器
	respErr := make(map[string]interface{})
	if userName == nil {
		//赋值
		respErr["status"] = 400
		respErr["msg"] = "当前用户未登录"
		//指定返回容器
		this.Data["json"] = respErr
		//指定返回方式
		this.ServeJSON()
		return
	}
	count, err1 := this.GetInt("count")
	fmt.Println("err1:=",err1)
	skuid, err2 := this.GetInt("skuid")
	fmt.Println("err2:=",err2)
	if err1 != nil || err2 != nil {
		//赋值
		respErr["status"]=401
		respErr["msg"]="获取数据异常"
		//指定返回容器
		this.Data["json"]=respErr
		//指定返回方式
		this.ServeJSON()
		return
	}
	//处理数据
	//把数据添加到redis的hash中
	err:=models.AddCart(userName.(string),skuid,count)
	if err!=nil{
		//赋值
		respErr["status"]=402
		respErr["msg"]="添加购物车失败"
		//指定返回容器
		this.Data["json"]=respErr
		//指定返回方式
		this.ServeJSON()
		return
	}
	//返回数据   ajax发送的是json数据,返回给ajax也得是json数据
	//1. 定义一个容器
	resp:=make(map[string]interface{})
	//2. 给容器赋值
	resp["status"]=200
	resp["errmsg"]="OK"
	//3. 指定容器返回
	this.Data["json"]=resp
	//4. 指定返回方式
	this.ServeJSON()

}

//展示购物车页面
func (this *CartController) ShowCart() {
	//从redis中获取购物车数据
	userName:=this.GetSession("userName")
	goods,totalCount,err:=models.GetCartData(userName.(string))
	if err!=nil{
		this.Data["errmsg"]="获取购物车数据失败"
		this.TplName="index_sx.html"
		return
	}
	this.Data["totalCount"]=totalCount
	this.Data["goods"]=goods
	this.TplName = "cart.html"
}

//删除购物车数据
func (this *CartController)DeleteCart () {
	resp:=make(map[string]interface{})
	//登录校验
	userName:=this.GetSession("userName")
	if userName==nil{
		resp["status"]=401
		resp["msg"]="当前用户未登录"
		//指定返回容器
		this.Data["json"]=resp
		this.ServeJSON()
		return
	}
	skuid,err:=this.GetInt("skuid")
	fmt.Println("err:=",err)
	if err!=nil{
		resp["status"]=402
		resp["msg"]="获取数据错误"
		this.Data["json"]=resp
		this.ServeJSON()
		return
	}
	//删除购物车数据
	err=models.DeleteCart(userName.(string),skuid)
	fmt.Println("err:=",err)
	if err!=nil{
		resp["status"]=403
		resp["msg"]="删除数据失败"
		return
	}
	resp["status"]=200
	resp["msg"]="OK"
	this.Data["json"]=resp
	this.ServeJSON()
}

