package api

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"fmt"
	"net/url"
	"net"
	"log"
	"strconv"
	"encoding/xml"
	"encoding/json"
	"github.com/runthinker/allpay/svc/model"
	"github.com/runthinker/allpay/wxpay"
	"github.com/runthinker/allpay/common"
	"github.com/runthinker/allpay/alipay"
)

type TWapPay struct {

}

var Wap TWapPay

func createWxAuthUrlForCode(reUrl string) string {
	authurl := &url.Values{}
	authurl.Set("appid",model.ApConfig.WxPay.Appid)
	authurl.Set("redirect_uri",reUrl)
	authurl.Set("response_type","code")
	authurl.Set("scope","snsapi_base")
	authurl.Set("state","STATE")
	return "https://open.weixin.qq.com/connect/oauth2/authorize?" + authurl.Encode() + "#wechat_redirect"
}

func createWxAuthUrlFormOpenid(code string) string  {
	authurl := &url.Values{}
	authurl.Set("appid",model.ApConfig.WxPay.Appid)
	authurl.Set("secret",model.ApConfig.WxPay.MpSecret)
	authurl.Set("code",code)
	authurl.Set("grant_type","authorization_code")
	return "https://api.weixin.qq.com/sns/oauth2/access_token?" + authurl.Encode()
}

func getOpenidFromMp(code string) string  {
	h := &http.Client{}
	geturl := createWxAuthUrlFormOpenid(code)
	resp,err := h.Get(geturl)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	respparams := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&respparams)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	return respparams["openid"].(string)
}

func (a *TWapPay)doWxPagePay(urlparams url.Values,openid string,ipaddr string,isSandBox bool) string  {
	mid := urlparams.Get("mid")
	subject := urlparams.Get("subject")
	total_amount := urlparams.Get("total_amount")
	out_trade_no := urlparams.Get("out_trade_no")
	return_url := urlparams.Get("return_url")

	mp,err := model.Wap.GetMchParams(mid,0)
	log.Println("mp=>",mp)
	if err != nil {
		log.Println("商户号后台未配置")
	}
	signkey := model.ApConfig.WxPay.Apikey
	attach := "Release"
	if isSandBox {
		signkey = model.ApConfig.WxPay.SandBoxSignkey
		attach = "SandBox"
	}
	cli := wxpay.NewWxClient(signkey,isSandBox)
	cli.FillDefaultParams()
	cli.PutMapParams(map[string]string{
		"appid": model.ApConfig.WxPay.Appid,
		"mch_id": model.ApConfig.WxPay.MchId,
	})
	cli.PutParam("notify_url",model.ApConfig.WxPay.NotifyUrl)
	if len(mp) > 0 {
		//设置了sub_mch_id
		cli.PutMapParams(mp)
	}
	cli.PutParam("attach",attach)
	cli.PutParam("openid",openid)
	cli.PutParam("body",subject)
	cli.PutParam("out_trade_no",out_trade_no)
	iamount,_ := strconv.Atoi(total_amount)
	cli.PutAnyParam("total_fee",iamount)
	cli.PutParam("spbill_create_ip",ipaddr)
	cli.PutParam("trade_type","JSAPI")
	//cli.PutParam("trade_type","NATIVE")
	//cli.PutParam("product_id","10000001")
	resp,err := cli.Execute("unifiedorder")
	log.Println(resp)
	if err != nil {
		log.Println("resp=>",err.Error())
		return common.BuildMsgPage(err.Error())
	}
	//return_url自带订单号，以便查询订单 支付宝不用查订单，url里自动返回了信息
	retp,_ :=url.ParseRequestURI(return_url)
	retp.Query().Set("out_trade_no",out_trade_no)
	retp.RawQuery = retp.Query().Encode()
	h5page := common.BuildwxjsCall(resp["appid"].(string),resp["prepay_id"].(string),retp.String(),signkey)
	log.Println(h5page)
	return h5page
}

func (a *TWapPay)doAliPagePay(urlparams url.Values,isSandBox bool) string  {
	mid := urlparams.Get("mid")
	subject := urlparams.Get("subject")
	total_amount := urlparams.Get("total_amount")
	out_trade_no := urlparams.Get("out_trade_no")
	return_url := urlparams.Get("return_url")

	mp,err := model.Wap.GetMchParams(mid,1)
	if err != nil {
		return common.BuildMsgPage("商户号后台未配置")
	}
	gateurl := ""
	if isSandBox {
		gateurl = common.K_ALI_PAY_SANDBOX_API_URL
	} else {
		gateurl = common.K_ALI_PAY_PRODUCTION_API_URL
	}
	cli := alipay.NewaliClient(model.ApConfig.AliPay.MchPrivatekey,model.ApConfig.AliPay.AliPublickey,gateurl)
	cli.FillDefaultParams()
	cli.PutSysParams("app_id",model.ApConfig.AliPay.AppId)
	cli.PutSysParams("notify_url",model.ApConfig.AliPay.NotifyUrl)
	cli.PutSysParams("return_url",return_url)
	if len(mp) > 0 {
		//设置了seller_id
		cli.PutMapAppParams(mp)
	}
	if model.ApConfig.AliPay.ProviderId != "" {
		cli.PutAppAnyParams("extend_params", map[string]string{
			"sys_service_provider_id": model.ApConfig.AliPay.ProviderId,
		})
	}
	cli.PutAppParams("subject", subject)
	cli.PutAppParams("out_trade_no", out_trade_no)
	iamount,_ := strconv.Atoi(total_amount)
	famount, _ := strconv.ParseFloat(fmt.Sprintf("%.2f",float64(iamount)/100),64)
	cli.PutAppAnyParams("total_amount",famount)
	resp,err := cli.PageExecute("alipay.trade.wap.pay")
	if err != nil {
		return common.BuildMsgPage(err.Error())
	}
	return string(resp)
}

func (a *TWapPay)DoPagePay(w http.ResponseWriter, r *http.Request, p httprouter.Params)  {
	log.Println("enter DoPagePay...")
	q := r.URL.Query()
	mid := q.Get("mid")
	method := q.Get("method")
	openid := q.Get("openid")
	if mid == "" || method == "" {
		fmt.Fprintln(w,common.BuildMsgPage("参数错误"))
		return
	}
	ipaddr := r.Header.Get("X-Real-IP")
	if len(ipaddr) == 0 {
		ipaddr, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	isSandBox := false
	if p.ByName("mode") == "sandbox" {
		isSandBox = true
	}

	var body string
	switch method {
	case "Wx.wappay":
		log.Println("request",r.RequestURI)
		if openid == "" {
			code := q.Get("code")
			if code == "" {
				scheme := "http://"
				if r.TLS != nil {
					scheme = "https://"
				}
				fullurl := scheme + r.Host + r.RequestURI
				mpurl := createWxAuthUrlForCode(fullurl)
				http.Redirect(w,r,mpurl,http.StatusFound)
				return
			} else {
				openid = getOpenidFromMp(code)
				log.Println("openid:",openid)
			}
		}
		if openid == "" {
			w.Write([]byte(common.BuildMsgPage("openid为空")))
			return
		}
		body = a.doWxPagePay(q,openid,ipaddr,isSandBox)
	case "Ali.wappay":
		body = a.doAliPagePay(q,isSandBox)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//WriteHeader一定放在Set后面
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func (a *TWapPay)notifyWxPay(w http.ResponseWriter, r *http.Request, p httprouter.Params)  {
	log.Println("notify:",r.RequestURI)
	params,err := wxpay.ParseResponse(r.Body)
	if err != nil {
		log.Println(err.Error())
	}
	isSandBox := params["attach"].(string) == "SandBox"
	signkey := model.ApConfig.WxPay.Apikey
	if isSandBox {
		signkey = model.ApConfig.WxPay.SandBoxSignkey
	}
	if err := wxpay.VerifySign(params,common.Md5Sign,signkey); err == nil {
		//验签成功
		//写库
		log.Println(params)
	}
	//reqbody,_ := ioutil.ReadAll(r.Body)
	//log.Println(string(reqbody))

	respParams := make(common.WxParams)
	respParams["return_code"] = common.Success
	respParams["return_msg"] = "OK"
	respbody,_ := xml.Marshal(respParams)
	w.Header().Set("Content-Type",common.BodyXMLType)
	w.WriteHeader(http.StatusOK)
	w.Write(respbody)
}

func (a *TWapPay)notifyAliPay(w http.ResponseWriter, r *http.Request, p httprouter.Params)  {
	//ParseForm后r.Form才有值，原值在r.PostForm和Url并合并，见源代码
	r.ParseForm()
	log.Println(r.Form)
	params := make(common.AliParams)
	for k,v := range r.Form {
		params[k] = v[0]
	}
	if params.VerifyAsyncAliSign(model.ApConfig.AliPay.AliPublickey) {
		log.Println("准备写库:",params)
	}
	w.Write([]byte("success"))
}

func (a *TWapPay)NotifyPay(w http.ResponseWriter, r *http.Request, p httprouter.Params)  {
	if p.ByName("paytype") == "wxpay" {
		a.notifyWxPay(w,r,p)
		return
	} else if p.ByName("paytype") == "alipay" {
		a.notifyAliPay(w,r,p)
		return
	}
}