package models

import (
	"github.com/garyburd/redigo/redis"
	"github.com/astaxie/beego/orm"
	"strconv"
)

//根据商品id获取商品数量和详细信息
func GetGoodsByld(skuids []string,userName string)([]map[string]interface{},int,int,error){
	//链接redis
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		return nil,0,0,err
	}
	defer conn.Close()
	o:=orm.NewOrm()
	//定义大容器
	var goods []map[string]interface{}
	//定义总金额变量
	var totalPrice,totalCount int

	for _,id:=range skuids{
		//先把字符串id转换成整形id
		idInt,err:=strconv.Atoi(id)
		if err!=nil{
			return nil,0,0,err
		}
		//获取商品数量
		count,err:=redis.Int(conn.Do("hget",userName+"_cart",idInt))
		if err!=nil{
			return nil,0,0,err
		}
		//获取商品详细信息
		var goodsSKU GoodsSKU
		goodsSKU.Id=idInt
		err=o.Read(&goodsSKU)
		if err!=nil{
			return nil,0,0,err
		}
		//定义一个行容器
		temp:=make(map[string]interface{})
		//赋值
		xiaoji:=goodsSKU.Price*count
		temp["count"]=count
		temp["xiaoji"]=xiaoji
		temp["goodsSKU"]=goodsSKU

		//把行容器追加到大容器中
		goods=append(goods,temp)
		//计算总价
		totalPrice+=xiaoji
		totalCount+=count
	}
	return goods,totalCount,totalPrice,err
}

//获取用户默认地址
func GetAddress (userName string)([]Address,error){
	o:=orm.NewOrm()
	var address []Address
	_,err:=o.QueryTable("Address").
		RelatedSel("User").
			Filter("User__Name",userName).
				All(&address)
	return address,err
}