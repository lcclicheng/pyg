package models

import (
	"github.com/astaxie/beego/orm"
	"errors"
	"github.com/garyburd/redigo/redis"
)

//用户注册
func Register(phone, pwd string) error {
	var user User
	o := orm.NewOrm()
	user.Name = phone
	user.PassWord = pwd
	_, err := o.Insert(&user)
	return err
}

//用户邮箱激活
func ActiveUser(userName, email string) error {
	//获取结构体对象
	var user User
	//获取orm对象
	o := orm.NewOrm()
	//给对象赋值查询条件
	user.Name = userName
	//查询更新用户是否存在
	if err := o.Read(&user, "Name"); err == nil {
		//更新操作
		//赋新值
		user.Email = email
		user.Active = true
		//更新
		_, err := o.Update(&user, "Email", "Active")
		return err
	} else {
		return err
	}
}

//用户登录
func Login(userName, pwd string) error {
	//获取结构体
	var user User
	//获取orm对象
	o := orm.NewOrm()
	//赋值
	user.Name = userName
	//查询
	if err := o.Read(&user, "Name"); err == nil {
		if user.PassWord != pwd {
			return errors.New("用户密码错误")
		}
		if !user.Active {
			return errors.New("当前用户未激活")
		}
		return nil
	} else {
		return errors.New("用户名输入错误")
	}
}

//用户地址
func AddSiter(receiver, detailAdd, zipCode, phone, userName string) error {
	//获取结构体对象
	var addr Address
	//获取orm对象
	o := orm.NewOrm()
	//根据用户名获取到用户对象
	var user User
	user.Name = userName
	o.Read(&user, "Name")
	addr.User = &user

	//在插入新地址(默认地址)之前,把原来的默认地址更新为非默认地址
	//判断当前数据库有默认地址
	//先校验是否有默认地址
	addr.IsDefault = 1
	err := o.QueryTable("Address").
		RelatedSel("User").
		Filter("User__Name", userName).Filter("IsDefault", 1).One(&addr)
	if err != nil {
		addr.Receiver = receiver
		addr.ZipCode = zipCode
		addr.Phone = phone
		addr.Addr = detailAdd
		addr.IsDefault = 1
		addr.User = &user
		o.Insert(&addr)
	} else {
		addr.IsDefault = 0
		o.Update(&addr, "IsDefault")

		//插入新地址
		var newaddr Address
		newaddr.Receiver = receiver
		newaddr.ZipCode = zipCode
		newaddr.Phone = phone
		newaddr.Addr = detailAdd
		newaddr.IsDefault = 1
		newaddr.User = &user
		o.Insert(&newaddr)
	}
	return err
}

//查询当前用户默认地址
func GetUserSite(userName string) (Address, error) {
	//定义orm对象
	o := orm.NewOrm()
	var address Address
	err := o.QueryTable("Address").
		RelatedSel("User").
		Filter("User__Name", userName).Filter("IsDefault", 1).One(&address)
	return address, err
}

//获取用户历史浏览记录
func GetUserHistory(userName string)([]GoodsSKU,error){
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		return nil,err
	}
	defer conn.Close()
	ids,err:=redis.Ints(conn.Do("lrange",userName+"_history",0,4))
	if err!=nil{
		return nil,err
	}
	o:=orm.NewOrm()
	var goodsSKUs []GoodsSKU
	for _,id:=range ids{
		var goods GoodsSKU
		goods.Id=id
		err:=o.Read(&goods,"Id")
		if err!=nil{
			return nil,err
		}
		goodsSKUs=append(goodsSKUs,goods)
	}
	return goodsSKUs,nil

}