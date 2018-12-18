# allpay
## Golang实现的聚合支付，含微信支付和支付宝  

* 特点  
1、代码简单、灵活可靠。  
2、可扩展性强：参数传递采用map[string]interface{}，不需要定义复杂的结构，满足多变的参数需求。  
3、支持服务商模式和非服务商模式。

* 第三方包引用  
1、httprouter  
2、gorm  

* 微信支付调用：  
```go
func TestWxClient_Execute(t *testing.T) {
	cli := NewWxClient("你的apikey",true)
	cli.FillDefaultParams()
	cli.PutMapParams(map[string]string{
		"appid": "您的appid",
		"mch_id": "您的商户号",
	})
	cli.PutParam("body","test")
	cli.PutParam("out_trade_no","58867657575757")
	cli.PutParam("sub_mch_id","您好的子商户号")
	cli.PutParam("total_fee","101")
	cli.PutParam("spbill_create_ip","127.0.0.1")
	cli.PutParam("notify_url","http://test.xxx.com/notify")
	cli.PutParam("trade_type","JSAPI")
	
	resp,err := cli.Execute("unifiedorder")
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log(resp)
}  
```

* 支付宝调用：  
```go
func doWapPay() []byte {
	cli := NewaliClient(con_privatekey,con_pubickey,"https://openapi.alipaydev.com/gateway.do")
	cli.FillDefaultParams()
	cli.PutMapSysParams(map[string]string{
		"app_id": "您的app_id",
		"notify_url": "http://test.xxx.com/notify",
	})
	cli.PutMapAppParams(map[string]string{
		"subject": "测试",
		"out_trade_no": "2000100001",
		"seller_id": "商户号",
	})
	cli.PutAppAnyParams("total_amount",1.00)
	cli.PutAppAnyParams("extend_params",map[string]string{
		"sys_service_provider_id": "服务商id",
	})
	resp,err := cli.PageExecute("alipay.trade.wap.pay")
	if err != nil {
		return []byte("")
	}
	return resp
}

func TestAliClient_Execute(t *testing.T) {
	h := httprouter.New()
	h.GET("/api/wap", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		body := doWapPay()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		//WriteHeader一定放在Set后面
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})
	http.ListenAndServe("localhost:3000",h)
}
```

* 聚合支付网页调用（javascript)  
```javascript
  function callpay(method) {
    let domain = location.href;
    let urlobj = {
        method: method,
        mid: "100000",
        subject: "测试商品",
        total_amount: "101",
        out_trade_no: new Date().Format("yyMMddhhmmss"),
        return_url: domain + "result.html?method=" + method,
    }

    let url = ""
    for(let k in urlobj) {
        if(url == "") {
            url = "?" + k + "=" + encodeURI(urlobj[k])
        } else {
            url += "&" + k + "=" + encodeURI(urlobj[k])
        }
    }
    let openurl = domain + "payapi/wappay/sandbox" + url;
    if(method === "Ali.wappay"){
        _AP.pay(openurl);    
    }else {
        location.href = openurl;
    }
  }  
```

* 使用到的数据库结构：  
```sql
/*==============================================================*/
/* Table: ap_mchpay_options                                     */
/*==============================================================*/
create table ap_mchpay_options
(
   mid                  bigint not null,
   paytype              tinyint not null,
   attrname             varchar(64) not null,
   attrvalue            varchar(128) not null,
   primary key (mid, paytype)
);

/*==============================================================*/
/* Table: ap_merchant_info                                      */
/*==============================================================*/
create table ap_merchant_info
(
   mid                  bigint not null auto_increment,
   name                 varchar(128) not null,
   isdebug              bool not null,
   apikey               varchar(128) not null,
   parentid             bigint not null default -1,
   mchclass             int not null default 0,
   primary key (mid)
)
auto_increment = 100000;
```
