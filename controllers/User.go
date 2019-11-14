package controllers

import (
	"github.com/astaxie/beego"
	"fmt"
	"math/rand"
	"time"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"pyg/models"
	"regexp"
	"github.com/astaxie/beego/utils"
)

type UserController struct {
	beego.Controller
}

func (this *UserController) ShowReg() {
	this.TplName = "register.html"
}
func (this *UserController) SendSms()  {
	//获取数据
	phone := this.GetString("phone")
	fmt.Println("phone:=", phone)

	//随机生成一个验证码  6位数的字符串
	rand.Seed(time.Now().UnixNano())

	//把整形转换成6位字符串,不足6位补0
	num := fmt.Sprintf("%06v", rand.Int31n(1000000))

	//调用阿里云接口发送短信
	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", "LTAI4Fw4e6bi2MpsWc1jorcr", "IXg1YZiM3dFGoHtqhpqnAtpVXVuR4M")

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"

	request.Domain = "dysmsapi.aliyuncs.com" //域名

	request.PhoneNumbers = phone
	request.SignName = "天天生鲜"
	request.TemplateCode = "SMS_174270423"
	request.TemplateParam = `{"code":` + num + `}`

	response, err := client.SendSms(request)
	fmt.Println(err, response)

	//获取一个容器   ajax传输的数据是json类型,接受的数据也是json类型
	resp := make(map[string]interface{})
	if response.IsSuccess() {
		//如果短信发送成功,把验证码存储到
		this.Ctx.SetCookie(phone+"_smsCode", num, 60*5)
		//成功状态返回前端   ajax

		//赋值
		resp["status"] = 200
		resp["msg"] = "OK"
		this.Data["json"] = resp
		//指定返回方式
		this.ServeJSON()
	} else {
		resp["status"] = 500
		resp["errmsg"] = err
		//返回数据
		this.Data["json"] = resp
		this.ServeJSON()
	}

}

func (this *UserController) HandleReg() {
	//获取数据
	phone := this.GetString("phone")
	code := this.GetString("code")
	pwd := this.GetString("password")
	rpwd := this.GetString("repassword")
	//校验数据
	if phone == "" || code == "" || pwd == "" || rpwd == "" {
		this.Data["errmsg"] = "输入数据不完整"
		this.TplName = "register.html"
		return
	}

	//校验验证码
	smsCode := this.Ctx.GetCookie(phone + "_smsCode")
	if code != smsCode {
		this.Data["errmsg"] = "验证码错误"
		this.TplName = "register.html"
		return
	}
	//校验两次密码是否相同
	if pwd != rpwd {
		this.Data["errmsg"] = "输入密码不一致"
		this.TplName = "register.html"
		return
	}
	//把数据插入数据库
	err := models.Register(phone, pwd)
	if err != nil {
		this.Data["errmsg"] = "用户注册失败"
		this.TplName = "register.html"
		return
	}
	//跳转到邮件页面
	this.Redirect("/registerEmail?userName="+phone, 302)
}
func (this *UserController) ShowEmail() {
	//获取用户名
	userName := this.GetString("userName")

	this.Data["userName"] = userName
	this.TplName = "register-email.html"
}
func (this *UserController) HandleActive() {
	//获取数据  获取邮箱  获取激活用户名
	email := this.GetString("email")
	userName := this.GetString("userName")

	//校验数据
	if email == "" || userName == "" {
		this.Data["errmsg"] = "输入数据不完整呢"
		this.TplName = "register-email.html"
		return
	}

	//邮箱地址匹配
	reg, err := regexp.Compile(`^[_a-z0-9-]+(\.[_a-z0-9-]+)*@[a-z0-9-]+(\.[a-z0-9-]+)*(\.[a-z]{2,})$`)
	if err != nil {
		this.Data["errmsg"] = "正则表达式错误"
		this.TplName = "register-email.html"
		return
	}
	//用规则去匹配需要匹配的字符串
	result := reg.MatchString(email)
	if !result {
		this.Data["errmsg"] = "邮箱格式不正确"
		this.TplName = "register-email.html"
		return
	}
	//发送邮件
	config := `{"username":"1210233414@qq.com","password":"rclybqmcfqctgebh","host":"smtp.qq.com","port":587}`
	emailSend := utils.NewEMail(config)

	emailSend.From = "1210233414@qq.com"
	emailSend.To = []string{email}
	emailSend.Subject = "品优购用户激活"
	ip := beego.AppConfig.String("ip")
	port := beego.AppConfig.String("httpport")
	emailSend.HTML = `<a href="http://` + ip + `:` + port + `/login?userName=` + userName + `&email=` + email + `">点击激活该用户</a>`
	err = emailSend.Send()
	fmt.Println("errmsg=", err)
	this.Ctx.WriteString("激活成功")
}

//获取数据
func (this *UserController) ActiveUser() {
	//获取数据
	userName := this.GetString("userName")
	email := this.GetString("email")
	if userName == "" || email == "" {
		this.Data["errmsg"] = "邮箱激活错误"
		this.TplName = "register-email.html"
		return
	}

	//处理数据   激活用户  在数据库做更新操作
	err := models.ActiveUser(userName, email)
	if err != nil {
		this.Data["errmsg"] = "激活失败"
		this.TplName = "register-email.html"
		return
	}
	//跳转登录
	this.Redirect("/login", 302)
}

//登录页面
func (this *UserController) ShowLogin() {
	this.TplName = "login.html"
}
//处理登录业务
func (this *UserController) HandleLogin() {
	//获取数据
	userName := this.GetString("userName")
	pwd := this.GetString("password")
	//校验数据
	if userName == "" || pwd == "" {
		this.Data["errmsg"]="输入数据不完整"
		this.TplName="login.html"
		return
	}
	//处理数据
	err:=models.Login(userName,pwd)
	fmt.Println("err:=",err)
	if err!=nil{
		this.Data["errmsg"]=err
		this.TplName="login.html"
		return
	}
	this.SetSession("userName",userName)
	this.Redirect("/index",302)
}

//退出登录
func (this *UserController)ShowLogout(){
    this.DelSession("userName")
    this.Redirect("/index.html",302)
}


//展示用户中心信息
func (this *UserController)ShowUserInfo(){
	//校验登录状态
	//获取用户名
	userName:=this.GetSession("userName")
	//获取当前地址的默认 地址
	address,err:=models.GetUserSite(userName.(string))
	if err!=nil{
		this.Data["addr"]=""
	}else {
		this.Data["addr"]=address
	}
	goodsSKUs,err:=models.GetUserHistory(userName.(string))
	if err!=nil{
		this.Data["goodsSKUs"]=""
	}else{
		this.Data["goodsSKUs"]=goodsSKUs
	}
	this.Data["userName"]=userName.(string)
	this.Layout="userCenterLayout.html"
	this.TplName="user_center_info.html"
}

func (this *UserController)ShowUserSite(){
	//获取当前用户默认地址  去数据库中查询
	userName:=this.GetSession("userName")
	address,err:=models.GetUserSite(userName.(string))
	address.Phone=address.Phone[:3]+"****"+address.Phone[7:]
	if err!=nil{
		this.Data["addr"]=""
	}else {
		this.Data["addr"]=address
	}
	this.Layout="userCenterLayout.html"
	this.TplName="user_center_site.html"
}
func (this *UserController)HandleSite(){
	//获取数据
	receiver:=this.GetString("receiver")
	detailAddr:=this.GetString("detailAddr")
	zipCode:=this.GetString("zipCode")
	phone:=this.GetString("phone")
	userName:=this.GetSession("userName")
	if receiver==""||detailAddr==""||zipCode==""||phone==""{
		this.Data["errmsg"]="数据不完整"
		this.Layout="userCenterLayout.html"
		this.TplName="user_center_site.html"
		return
	}
	//处理数据   插入地址数据
	err:=models.AddSiter(receiver,detailAddr,zipCode,phone,userName.(string))
	if err != nil {
		this.Data["errmsg"]="添加地址失败"
		this.Layout="userCenterLayout.html"
		this.TplName="user_center_site.html"
		return
	}
	//返回数据
	this.Redirect("/user/userCenterSite",302)
}
func (this *UserController)ShowUserOrder(){
	this.Layout="userCenterLayout.html"
	this.TplName="user_center_order.html"
}