package routers

import (
	"pyg/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {

	filterFunc := func(ctx *context.Context) {
		userName := ctx.Input.Session("userName")
		if userName == nil {
			ctx.Redirect(302, "/login")
			return
		}
	}
	//路由过滤器
	beego.InsertFilter("/user/*", beego.BeforeExec, filterFunc)
	//注册功能路由
	beego.Router("/register", &controllers.UserController{}, "get:ShowReg;post:HandleReg")
	//发送验证码路由
	beego.Router("/sendSms", &controllers.UserController{}, "post:SendSms")
	//邮箱注册路由
	beego.Router("/registerEmail", &controllers.UserController{}, "get:ShowEmail;post:HandleActive")
	//激活用户路由
	beego.Router("/activeUser", &controllers.UserController{}, "get:ActiveUser")
	//登录路由
	beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")
	//主页面路由
	beego.Router("/index", &controllers.GoodsController{}, "get:ShowIndex")
	//退出登录路由
	beego.Router("/logout", &controllers.UserController{}, "get:ShowLogout")
	//用户中心路由
	beego.Router("/user/userCenterInfo", &controllers.UserController{}, "get:ShowUserInfo")
	//用户收货地址路由
	beego.Router("/user/userCenterSite", &controllers.UserController{}, "get:ShowUserSite;post:HandleSite")
	//用户订单路由
	beego.Router("/user/userCenterOrder", &controllers.UserController{}, "get:ShowUserOrder")
	//生鲜主页面路由
	beego.Router("/indexSx", &controllers.GoodsController{}, "get:ShowIndexSx")
	//商品详情路由
	beego.Router("/detail", &controllers.GoodsController{}, "get:ShowDetail")
	//商品列表路由
	beego.Router("/list", &controllers.GoodsController{}, "get:ShowList")
	//搜索商品路由
	beego.Router("/searchGoods", &controllers.GoodsController{}, "post:SearchGoods")
	//添加购物车路由
	beego.Router("/addCart", &controllers.CartController{}, "post:AddCart")
	//我的购物车路由
	beego.Router("/user/showCart", &controllers.CartController{}, "get:ShowCart")
	//删除购物车路由
	beego.Router("/deleteCart", &controllers.CartController{}, "post:DeleteCart")
	//去结算
	beego.Router("/jiesuan",&controllers.OrderController{},"get:ShowOrder")
	//插入订单数据
	beego.Router("/insertOrder",&controllers.OrderController{},"post:InsertOrder")
	//进入支付页面
	beego.Router("/pay",&controllers.OrderController{},"get:PayAli")
}
