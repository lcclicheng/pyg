package controllers

import (
	"github.com/astaxie/beego"
	"pyg/models"
	"github.com/astaxie/beego/orm"
	"time"
	"strconv"
	"github.com/garyburd/redigo/redis"
	"strings"
	"fmt"
	"github.com/smartwalle/alipay"
)

type OrderController struct {
	beego.Controller
}

//展示支付界面
func (this *OrderController) ShowOrder() {
	//获取数据
	skuids := this.GetStrings("skuid")
	//校验数据
	if len(skuids) == 0 {
		this.Redirect("/user/showCart", 302)
		return
	}
	//获取用户名
	userName := this.GetSession("userName")
	if userName == nil {
		this.Redirect("/login", 302)
		return
	}
	//处理数据
	goods, totalCount, totalPrice, err := models.GetGoodsByld(skuids, userName.(string))
	if err != nil {
		this.Data["errmsg"] = "获取用户信息失败"
		this.TplName = "place_order.html"
		return
	}

	//获取用户地址
	userName = this.GetSession("userName")
	address, err := models.GetAddress(userName.(string))
	if err != nil {
		this.Data["address"] = ""
	} else {
		this.Data["address"] = address
	}
	this.Data["totalCount"] = totalCount
	this.Data["totalPrice"] = totalPrice
	this.Data["totalTruePrice"] = totalPrice + 10
	this.Data["goods"] = goods
	this.Data["skuids"] = skuids
	this.TplName = "place_order.html"
}

//插入订单数据
func (this *OrderController) InsertOrder() {
	resp := make(map[string]interface{})
	//从sessin中获取用户名
	userName := this.GetSession("userName")
	if userName == nil {
		resp["status"] = 400
		resp["msg"] = "当前用户登录已失效"
		this.Data["json"] = resp
		this.ServeJSON()
		return
	}

	//获取数据
	addrId, err1 := this.GetInt("addrId")
	payId, err2 := this.GetInt("payId")
	skuids := this.GetString("skuids")
	totalCount, err3 := this.GetInt("totalCount")
	totalPrice, err4 := this.GetInt("totalPrice")
	transit, err5 := this.GetInt("transit")
	fmt.Println("addrId=", addrId, "payId=", payId, "skuids=", skuids, "count=", totalCount, "Price=", totalPrice, "transit=", transit)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		resp["status"] = 401
		resp["msg"] = "获取数据错误"
		this.Data["json"] = resp
		this.ServeJSON()
		return
	}
	//插入数据到订单表
	var orderInfo models.OrderInfo
	o := orm.NewOrm()
	//根据用户名获取用户对象
	var user models.User
	user.Name = userName.(string)
	err := o.Read(&user, "Name")
	if err != nil {
		resp["status"] = 402
		resp["mag"] = "获取用户信息错误"
		this.Data["json"] = resp
		this.ServeJSON()
		return
	}
	orderInfo.User = &user
	orderInfo.TotalCount = totalCount
	orderInfo.TotalPrice = totalPrice
	orderInfo.TransitPrice = transit
	orderInfo.PayMethod = payId

	//获取订单号
	orderId := time.Now().Format("20060102150405") + strconv.Itoa(user.Id)
	orderInfo.OrderId = orderId

	//获取地址
	var addr models.Address
	addr.Id = addrId
	o.Read(&addr)

	orderInfo.Address = &addr
	//在插入订单表之前开启事务
	o.Begin()

	_, err = o.Insert(&orderInfo)
	if err != nil {
		o.Rollback()
		resp["status"] = 403
		resp["msg"] = "插入订单失败"
		this.Data["json"] = resp
		this.ServeJSON()
		return
	}
	//连接redis
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		o.Rollback()
		resp["status"] = 404
		resp["msg"] = "redis链接失败"
		this.Data["json"] = resp
		this.ServeJSON()
		return
	}
	defer conn.Close()

	//插入订单商品
	ids := strings.Split(skuids[1:len(skuids)-1], " ")
	for _, id := range ids {
		var goodsSKU models.GoodsSKU
		idInt, _ := strconv.Atoi(id)
		goodsSKU.Id = idInt
		o.Read(&goodsSKU)
		//存表示量
		historyStock:=goodsSKU.Stock
		var orderGoods models.OrderGoods
		orderGoods.OrderInfo = &orderInfo
		orderGoods.GoodsSKU = &goodsSKU

		//从redis中获取商品数量
		count, err := redis.Int(conn.Do("hget", userName.(string)+"_cart", idInt))
		if err != nil {
			o.Rollback()
			resp["status"] = 405
			resp["msg"] = "获取商品数量失败"
			this.Data["json"] = resp
			this.ServeJSON()
			return
		}
		orderGoods.Count = count
		orderGoods.Price = goodsSKU.Price * count
		if goodsSKU.Stock < count {
			o.Rollback()
			resp["status"] = 406
			resp["mag"] = "商品库存不足"
			this.Data["json"] = resp
			this.ServeJSON()
			return
		}
		_, err = o.Insert(&orderGoods)
		if err != nil {
			o.Rollback()
			resp["status"] = 407
			resp["msg"] = "插入订单失败"
			this.Data["json"] = resp
			this.ServeJSON()
			return
		}
		//插入订单商品成功,把商品从购物车删除,并在库存中减去购买量,销量增加购买量
		conn.Do("hdel", userName.(string)+"_cart", idInt)
		o.Read(&goodsSKU)
		fmt.Println("历史库存为:",historyStock,"现在库存为:",goodsSKU.Stock)
		if goodsSKU.Stock!=historyStock{
			resp["status"]=408
			resp["msg"]="库存不足"
			this.Data["json"]=resp
			this.ServeJSON()
			o.Rollback()
			return
		}
		goodsSKU.Stock -= count
		goodsSKU.Sales += count
		o.Update(&goodsSKU, "Stock", "Sales")
	}
	o.Commit()
	resp["status"] = 200
	resp["msg"] = "OK"
	resp["orderId"] = orderInfo.Id
	this.Data["json"] = resp
	this.ServeJSON()

}

//支付
func (this *OrderController) PayAli() {
	//生成阿里操作对象
	appid := "2016101200668434"
	publicKey := `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5Ddmek126VGWupsiw3Sc
uQp9wCbV/QQ3RUYTw74fTmYFqRpBt6gPgIxflCZaKV6edT6aeZVIdMK0RS+dlXZL
ND1Tcir2qIezWmMtc4DjDG+TOuy/8gcWeKNHPKNQ45949YhyoetkQdZ0COEVEceY
4jrVoXE2eNWJuXrmJkoyrelxVOcWyaDB0VrnlfbBYYkn8XUjCviAoBhNGqlqx7wE
8pWJ2qVEbydr4wxaJKxL9im+mkESU5IkKIy8jXBwbB5kyY6vtD3URElVYcPJhZJA
wnTYhrZ0+/GpREhL0dQTkTn6snK3M3HSZOqqZhlqikf6SPJzGiDzoEYT5GbCx0ys
6QIDAQAB
`
	privateKey := `MIIEowIBAAKCAQEA5Ddmek126VGWupsiw3ScuQp9wCbV/QQ3RUYTw74fTmYFqRpB
	t6gPgIxflCZaKV6edT6aeZVIdMK0RS+dlXZLND1Tcir2qIezWmMtc4DjDG+TOuy/
		8gcWeKNHPKNQ45949YhyoetkQdZ0COEVEceY4jrVoXE2eNWJuXrmJkoyrelxVOcW
	yaDB0VrnlfbBYYkn8XUjCviAoBhNGqlqx7wE8pWJ2qVEbydr4wxaJKxL9im+mkES
	U5IkKIy8jXBwbB5kyY6vtD3URElVYcPJhZJAwnTYhrZ0+/GpREhL0dQTkTn6snK3
	M3HSZOqqZhlqikf6SPJzGiDzoEYT5GbCx0ys6QIDAQABAoIBAQC94j6Y6lVTQnh4
	YVYmbKNt7wW8WFPaBqT6NZmCV3Fy6L4y+k7Nwb7MRX/NI7AHFdwgT2t2WDiGNe6K
	Vlj2oAtottH0fzzl8qrPPQ/3N7kygq9s6sm2ViFjVO+Ty4slKW4aVWKTyOiNQyMe
	tDC0r29MZImVnz4kgf/q3RAbscbDHZNE8qdcxPR1m8dmmVfrCl0knLiPmIGdy10s
	b7MdpcccGN/EJ/CcCE7M+OduhUsu3JgAgtcxDBcIno4mKQOuUEDH/JWjyBSxKvum
	jVoICJI8l6daOL2a2a80BrWjrOvJ4gSmNcWp0xSP5I1RqZkG7f1zXcLeInLVeBo9
	xSVVL99VAoGBAP86HOSPoL64ziO4YWLHdZ2fEb4AnjweECXTX/5A1uBEEH1nH9sS
	tA+VVOrQnsjDHLHh8TsbitD8ia+gyrmKaia02EC4Gto0k+MlgETV3qarH4uGD1g8
	Z6dWledKB9jPQcM+GQmXm26mH1JGLhdHursVD34AlFlgolJ+G1D7ZsFPAoGBAOTo
	WFjZKPMBHNBmFaMWSeeXsCctHe4ARhT/f/9wJ3syeG4G25u3lxagovfz89o2D7Ui
	XMCuDaTF0Yuh4MLQbX3/sUj/ewiz/5T1v6PaXCRkm/sCY6/nAHXj4p5laSnWt3M2
	f+wUN3wQcK7UvOeul4k2t2wCR9MZAvYu6eMbffBHAoGALbFkvNKt75c8aI64+KtG
	9kolLgQEUDT9pRf7ppRLI+lrnlfZDyqBDA0rH8Lrunub5ojR3EgpCRM9PzElOiR6
	rqVP1f5f6FLjaxYqqag0bVhTlHISyzQ9Rmss+TR6xSkN1/uFFf+LdzrMfrlLxSU1
	XAsANAm8hWfUh7pF/7CSi+cCgYBSUJcMhDLsh7bj3gHj3qz+4hZPUDMWFfUdse9G
	XP9llvWlo0OvkGp9kZBpF8nV62DkoxG1nCF94kEDNFgN1kO5boxDEtQcghXjbCfY
	9TnzQFazAd31MF2DB0rD6PXTPMpFXRDNRUvailLrG8c+jRMjHZEB/yPy7miZPK+Q
	op88GwKBgFb9US/UJ6h3SIWWvEGXT+273OI6ARUkOhFHfTmwnQEqnALqh6j/rqEJ
	HuEXTI8KHU634GgvGdRzGEx9IhMgRZa2FB+2wBcuPfaAiJE+4AyaTL1ju6VO3US9
	WOLTle0v0dUb3eRslvFYbF3GdtxHnBM1VrXPrOjO5Fg45n3Ph046`

	client, err := alipay.New(appid, publicKey, privateKey, false)
	if err != nil {
		this.Data["errmsg"] = "支付失败"
		this.TplName = "place_order.html"
		return
	}
	//支付配置
	p := alipay.TradePagePay{}
	//设置支付主题
	p.Subject = "品优购订单支付"
	//设置订单号
	p.OutTradeNo = "20190926123456"
	//设置支付金额
	p.TotalAmount = "1000.00"
	//设置异步返回方式
	p.NotifyURL = "http://127.0.0.1:8080/user/userCenterInfo"
	//设置同步返回方式
	p.ReturnURL = "http://127.0.0.1:8080/user/userCenterInfo"

	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	url, err :=client.TradePagePay(p)
	if err!=nil{
		this.Data["errmsg"]="支付失败"
		this.TplName="place_order.html"
		return
	}
	this.Redirect(url.String(),302)
}
