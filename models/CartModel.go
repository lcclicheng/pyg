package models

import (
	"github.com/garyburd/redigo/redis"
	"github.com/astaxie/beego/orm"
	"strconv"
)

func AddCart(userName string,goodsId,count int)error{
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if  err!=nil{
		return err
	}
	defer conn.Close()
	//获取原来数据库中的数据累加
	preCount,_:=redis.Int(conn.Do("hget",userName+"_cart",goodsId))
	_,err=conn.Do("hset",userName+"_cart",goodsId,count+preCount)
	return err
}

func GetCartData(userName string)([]map[string]interface{},int,error){
	var goods []map[string]interface{}
	var totalCount int
	//连接redis
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		return nil,0,err
	}
	defer conn.Close()
	//操作redis获取数据   id,count
	idMap,err:=redis.IntMap(conn.Do("hgetall",userName+"_cart"))
	if err!=nil{
		return nil,0,err
	}
	o:=orm.NewOrm()
	for id,count:=range idMap{
		//定义行容器
		temp:=make(map[string]interface{})
		//把id转成int类型
		idInt,err:=strconv.Atoi(id)
		if err!=nil{
			return nil,0,err
		}
		//根据id获取商品数据
		var goodsSKU GoodsSKU
		goodsSKU.Id=idInt
		err=o.Read(&goodsSKU)
		if err!=nil{
			return nil,0,err
		}
		//计算小计
		littleCount:=goodsSKU.Price*count

		//给行容器赋值

		temp["goodsSKU"]=goodsSKU
		temp["count"]=count
		temp["littleCount"]=littleCount
		goods=append(goods,temp)
		totalCount+=count
	}
	return goods,totalCount,err

}

func DeleteCart(userName string,skuid int)error{
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		return err
	}
	defer conn.Close()
	_,err=conn.Do("hdel",userName+"_cart",skuid)
	return err

}