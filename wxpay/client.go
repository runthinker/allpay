package wxpay

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"net/http"
	"strings"
	"io/ioutil"
	"log"
	"io"
	"github.com/runthinker/allpay/common"
	"github.com/runthinker/allpay/util"
)

type wxClient struct {
	apiKey    string
	isSandbox bool
	bodytype  string
	appParams common.WxParams
}

func NewWxClient(apikey string,isSandBox bool) *wxClient  {
	return &wxClient{apiKey:apikey,isSandbox:isSandBox,bodytype:common.BodyXMLType,appParams:make(common.WxParams)}
}

func (t *wxClient)PutParam(key string,value interface{})  {
	t.appParams[key] = value
}

func (t *wxClient)PutMapParams(mp map[string]string)  {
	for k,v := range mp {
		t.appParams[k] = v
	}
}

func (t *wxClient)PutAnyParam(key string,value interface{})  {
	t.appParams[key] = value
}

func (t *wxClient)HasParam(key string) bool  {
	_,has := t.appParams[key]
	return has
}

func (t *wxClient)FillDefaultParams()  {
	if _,has := t.appParams["nonce_str"]; !has {
		t.appParams["nonce_str"] = util.NonceStr(16)
	}
	if _,has := t.appParams["sign_type"]; !has {
		t.appParams["sign_type"] = common.Md5Sign
	}
}

func (t *wxClient)postData(h *http.Client,url string)(common.WxParams,error)  {
	//h := &http.Client{}
	if !t.HasParam("sign_type") {
		return nil,errors.New("无签名类型")
	}
	//签名
	signtype := t.appParams["sign_type"].(string)
	sign := t.appParams.MakeWxSign(signtype,t.apiKey)
	t.PutParam(common.WxSign,sign)

	body,err := xml.Marshal(t.appParams)
	log.Println("body:",string(body))
	if err != nil {
		return nil,err
	}
	log.Println("url:",url)
	resp,err := h.Post(url,t.bodytype,strings.NewReader(string(body)))
	if err != nil {
		return nil,err
	}
	//res, err := ioutil.ReadAll(resp.Body)
	//log.Println(string(res))
	respParams,err := ParseResponse(resp.Body)
	if err == nil {
		err = VerifySign(respParams,signtype,t.apiKey)
	}
	return respParams,err
}

func (t *wxClient)Execute(method string)(common.WxParams,error)  {
	var url string
	if t.isSandbox {
		//strings.Replace(url,`https://api.mch.weixin.qq.com/`,`https://api.mch.weixin.qq.com/sandboxnew/`,-1)
		url = `https://api.mch.weixin.qq.com/sandboxnew/pay/`
	} else {
		url = `https://api.mch.weixin.qq.com/pay/`
	}
	url += method
	h := &http.Client{}
	return t.postData(h,url)
}

func (t *wxClient)ExecuteSSL(method string,cert *tls.Certificate)(common.WxParams,error)  {
	var url string
	if t.isSandbox {
		url = `https://api.mch.weixin.qq.com/sandboxnew/pay/`
	} else {
		url = `https://api.mch.weixin.qq.com/pay/`
	}
	url += method
	h := util.NewHttpsClient(cert)
	return t.postData(h,url)
}

//解析Response，同时验签
func VerifySign(respParams common.WxParams,signtype string,apikey string) error  {
	//resp含有sign_type，以resp的sign_type为准
	if _,has := respParams["sign_type"]; has {
		signtype = respParams["sign_type"].(string)
	}

	if code,has := respParams["return_code"]; has {
		if code == common.Success {
			if ywcode,has2 := respParams["result_code"]; has2 && ywcode != common.Success {
				return errors.New(respParams["err_code_des"].(string))
			}
			if respParams.VerifyWxSign(signtype,apikey) {
				return nil
			}else {
				return errors.New("签名效验出错误")
			}
		} else {
			return errors.New(respParams["return_msg"].(string))
		}
	}
	return errors.New("respone not has return_code")
}

func ParseResponse(respBody io.ReadCloser)(common.WxParams,error)  {
	var respParams common.WxParams
	if err := xml.NewDecoder(respBody).Decode(&respParams); err != nil {
		return nil,err
	}
	//没有sign，不验签
	if _,has := respParams[common.WxSign]; !has {
		if code,has := respParams["return_code"]; has && code == common.Success {
			return respParams, nil
		} else {
			return nil,errors.New(respParams["return_msg"].(string))
		}
	}
	return respParams,nil
}

func GetSandBoxSignKey(mch_id string,apikey string) string  {
	h := &http.Client{}
	params := make(common.WxParams)
	params["mch_id"] = mch_id
	params["nonce_str"] = util.NonceStr(32)
	params["sign"] = params.MakeWxSign(common.Md5Sign,apikey)
	body,_ := xml.MarshalIndent(params,""," ")
	resp, err := h.Post("https://api.mch.weixin.qq.com/sandboxnew/pay/getsignkey",common.BodyXMLType,strings.NewReader(string(body)))
	defer resp.Body.Close()
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	res, err := ioutil.ReadAll(resp.Body)
	log.Println(string(res))
	var respMap common.WxParams
	err = xml.Unmarshal(res,&respMap)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	return respMap["sandbox_signkey"].(string)
}