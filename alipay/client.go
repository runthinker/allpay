package alipay

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"github.com/runthinker/allpay/common"
)

type aliClient struct {
	privateKey      string
	aliPayPublicKey string
	payGateway		string
	sysParams		common.AliParams
	appParams		common.AliParams
}

func NewaliClient(privkey,pubkey,url string)*aliClient  {
	return &aliClient{
		privateKey:privkey,
		aliPayPublicKey:pubkey,
		payGateway:url,
		sysParams:make(common.AliParams),
		appParams:make(common.AliParams),
	}
}

func (a *aliClient)PutMapSysParams(mp map[string]string)  {
	for k,v := range mp {
		a.sysParams[k] = v
	}
}

func (a *aliClient)PutSysParams(key,value string)  {
	a.sysParams[key] = value
}

func (a *aliClient)PutMapAppParams(mp map[string]string)  {
	for k,v := range mp {
		a.appParams[k] = v
	}
}

func (a *aliClient)PutAppParams(key,value string)  {
	a.appParams[key] = value
}

func (a *aliClient)PutAppAnyParams(key string,value interface{})  {
	a.appParams[key] = value
}


func (a *aliClient)FillDefaultParams()  {
	if _,has := a.sysParams["charset"]; !has {
		a.sysParams["charset"] = "utf-8"
	}
	if _,has := a.sysParams["format"]; !has {
		a.sysParams["format"] = "JSON"
	}
	if _,has := a.sysParams["sign_type"]; !has {
		a.sysParams["sign_type"] = "RSA2"
	}
	if _,has := a.sysParams["timestamp"]; !has {
		a.sysParams["timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	}
	if _,has := a.sysParams["version"]; !has {
		a.sysParams["version"] = "1.0"
	}
	if _,has := a.appParams["product_code"]; !has {
		a.appParams["product_code"] = "QUICK_WAP_WAY"
	}
}

func (a *aliClient)getRequestUrl()(string,error)  {
	allParams := a.sysParams.AppendSubParams("biz_content",a.appParams)
	log.Println(allParams)
	sign := allParams.MakeAliSign(a.privateKey)
	if len(sign) == 0 {
		return "",errors.New("签名出错")
	}
	log.Println("sign:",sign)
	a.PutSysParams(common.AliSign,sign)
	url := a.sysParams.GenUrlStr()
	var fullurl string
	if len(url) > 0 {
		fullurl = a.payGateway + "?" + url
	} else {
		fullurl = a.payGateway
	}
	return fullurl,nil
}

func (a *aliClient)getForm()(string,error)  {
	if _,has := a.sysParams["method"]; !has {
		return "",errors.New("method参数未设置")
	}
	baseurl,err := a.getRequestUrl()
	if err != nil {
		return "",err
	}
	formParams := make(map[string]string)
	js,_ := json.Marshal(a.appParams)
	//log.Println("biz_content=",url.QueryEscape(string(js)))
	formParams["biz_content"] = string(js)
	return common.Buildaliform(baseurl,formParams),nil
}

func (a *aliClient)post(h *http.Client)([]byte,error)  {
	if _,has := a.sysParams["method"]; !has {
		return nil,errors.New("method参数未设置")
	}
	baseurl,err := a.getRequestUrl()
	if err != nil {
		return nil,err
	}
	js,_ := json.Marshal(a.appParams)
	formdata := "biz_content=" + url.QueryEscape(string(js))
	log.Println(formdata)
	req,_ := http.NewRequest("POST",baseurl,strings.NewReader(formdata))
	req.Header.Set("Content-Type",common.BodyUrlencodedType)
	resp,err := h.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil,err
	}
	res, err := ioutil.ReadAll(resp.Body)
	return res,err
}

func (a *aliClient)postCheck(h *http.Client)(common.AliParams,error)  {
	resp,err := a.post(h)
	if err != nil {
		return nil,err
	}
	var respParams common.WxParams
	if err := json.Unmarshal(resp,&respParams); err != nil {
		log.Println("decode error")
		return nil,err
	}
	if code,has := respParams["code"]; has {
		if code == common.ALISUCCESS {
			 if respParams.VerifySyncAliSign(a.aliPayPublicKey) {
			 	return respParams,nil
			 } else {
			 	return nil,errors.New("验签出错")
			 }
		} else {
			return nil,errors.New(respParams["msg"].(string))
		}
	}
	return nil,errors.New("resp code not found!")
}

func (a *aliClient)PageExecute(method string)([]byte,error)  {
	a.PutSysParams("method",method)
	bodyform,err := a.getForm()
	return []byte(bodyform),err
}