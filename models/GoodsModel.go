package models

import (
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
)

//获取所有菜单
func GetAllMenu() []map[string]interface{} {
	o := orm.NewOrm()
	//定义总容器
	var goodsTypes []map[string]interface{}
	//获取一级菜单
	var menus []TpshopCategory
	o.QueryTable("TpshopCategory").Filter("Pid", 0).All(&menus)
	//获取二级菜单
	for _, v := range menus {
		temp := make(map[string]interface{})
		var two []TpshopCategory
		o.QueryTable("TpshopCategory").Filter("Pid", v.Id).All(&two)
		//给map赋值
		temp["one"] = v
		temp["two"] = two
		//把map放到总容器中
		goodsTypes = append(goodsTypes, temp)
	}
	//获取三级菜单
	for _, v1 := range goodsTypes {
		//根据二级菜单获取三级菜单
		var TwoTemp []map[string]interface{}
		for _, v2 := range v1["two"].([]TpshopCategory) {
			temp := make(map[string]interface{})
			var three []TpshopCategory
			o.QueryTable("TpshopCategory").Filter("Pid", v2.Id).All(&three)
			//给map赋值
			temp["three"] = three
			temp["four"] = v2
			//给TwoTemp赋值
			TwoTemp = append(TwoTemp, temp)
		}
		//把获取到二级对象以及三级对象切片的容器放到大容器中
		v1["three"] = TwoTemp
	}
	return goodsTypes
}

//获取所有类型
func GetAllType() ([]GoodsType, error) {
	o := orm.NewOrm()
	var goodsTypes []GoodsType
	_, err := o.QueryTable("GoodsType").All(&goodsTypes)
	return goodsTypes, err
}

//获取所有轮播图
func GetAllBanner() ([]IndexGoodsBanner, error) {
	o := orm.NewOrm()
	var banners []IndexGoodsBanner
	_, err := o.QueryTable("IndexGoodsBanner").RelatedSel("GoodsSKU").OrderBy("Index").All(&banners)
	return banners, err
}

//获取促销商品
func GetAllPromotion() ([]IndexPromotionBanner, error) {
	o := orm.NewOrm()
	var promotions []IndexPromotionBanner
	_, err := o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&promotions)
	return promotions, err
}

//获取首页展示商品
func GetAllIndexGoods(types []GoodsType) []map[string]interface{} {
	var goods []map[string]interface{}
	o := orm.NewOrm()
	for _, v := range types {
		//申请容器
		temp := make(map[string]interface{})
		//获取首页展示商品
		var TextgoodsIndex []IndexTypeGoodsBanner
		var ImggoodsIndex []IndexTypeGoodsBanner
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSKU").
			Filter("GoodsType__Id", v.Id).OrderBy("Index").Filter("DisplayType", 0).All(&TextgoodsIndex)
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSKU").
			Filter("GoodsType__Id", v.Id).OrderBy("Index").Filter("DisplayType", 1).All(&ImggoodsIndex)
		//给容器赋值
		temp["goodsType"] = v
		temp["imgGoods"] = ImggoodsIndex
		temp["textGoods"] = TextgoodsIndex
		goods = append(goods, temp)
	}
	return goods
}

//获得详情页面的商品类型
func GetGoodsDetail(id int) (error, GoodsSKU) {
	o := orm.NewOrm()
	var goodsSKU GoodsSKU
	err := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("Id", id).One(&goodsSKU)
	return err, goodsSKU
}

//获取新品
func GetNewGoods(typeName string) ([]GoodsSKU, error) {
	o := orm.NewOrm()
	var newGoods []GoodsSKU
	_, err := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name", typeName).OrderBy("-Time").Limit(2, 0).All(&newGoods)
	return newGoods, err
}

//获取同一种类所有的商品
func GetTypeGoods(typeName, sort string) ([]GoodsSKU, error) {
	o := orm.NewOrm()
	var AllGoodsTypes []GoodsSKU
	var err error
	//根据排序方式获取排序后的数据
	if sort == "price" {
		_, err = o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name", typeName).OrderBy("Price").All(&AllGoodsTypes)
	} else if sort == "sales" {
		_, err = o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name", typeName).OrderBy("Sales").All(&AllGoodsTypes)
	} else {
		_, err = o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name", typeName).All(&AllGoodsTypes)
	}
	return AllGoodsTypes, err
}

//查询当前类型对应商品总数量
func GetGoodsCount(typeName string) (int64, error) {
	o := orm.NewOrm()
	return o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name", typeName).Count()
}
func GetPageIndexGoods(pageSize, pageIndex int, typeName,sort string) ([]GoodsSKU, error) {
	o := orm.NewOrm()
	//获取数据的起始位置
	start := pageSize * (pageIndex - 1)
	var goods []GoodsSKU
	 _,err:=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsTYpe__Name", typeName).Limit(pageSize, start).All(&goods)
	return goods,err
}

//存储浏览数据
func SaveHistory(userName string,goodsId int)error{

	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		return err
	}
	defer conn.Close()
	//插入前把相同数据删除
	conn.Do("lrem",userName+"_history",0,goodsId)

	_,err= conn.Do("lpush",userName+"_history",goodsId)
	return err
}


func GetAllGoods()([]GoodsSKU,error){
	o:=orm.NewOrm()
	var goods []GoodsSKU
	_,err:=o.QueryTable("GoodsSKU").All(&goods)
	return goods,err
}

func SearchGoods(goodsName string)([]GoodsSKU,error){
	o:=orm.NewOrm()
	var goods []GoodsSKU
	_,err:=o.QueryTable("GoodsSKU").Filter("Name__contains",goodsName).All(&goods)
	return goods,err
}