package common

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"sort"
	"strings"
	"encoding/base64"
	"net/url"
	"bytes"
	"strconv"
	"fmt"
	"github.com/runthinker/allpay/util"
)

type WxParams map[string]interface{}
type AliParams = WxParams

type xmlMapEntry struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

const (
	Md5Sign = "MD5"
	HMacSha256 = "HMAC-SHA256"
	WxSign = "sign"
	AliSign = "sign"
	SignType = "sign_type"
)

func (p WxParams)MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(p) == 0 {
		return nil
	}

	start.Name.Local = "xml"
	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range p {
		switch v.(type) {
		case string:
			e.Encode(xmlMapEntry{XMLName: xml.Name{Local: k}, Value: v.(string)})
		case int:
			e.Encode(xmlMapEntry{XMLName: xml.Name{Local: k}, Value: strconv.Itoa(v.(int))})
		case WxParams:
			cdata,err := json.Marshal(v)
			log.Println(string(cdata))
			if err == nil {
				log.Println("enter")
				err = e.Encode(struct {
					XMLName xml.Name
					Value string `xml:",cdata"`
				}{xml.Name{Local: k},string(cdata)})
				if err != nil {
					log.Println(err.Error())
				}
			}
		}
	}

	return e.EncodeToken(start.End())
	/*
	tokens := []xml.Token{start}

	for key, value := range p {
		t := xml.StartElement{Name: xml.Name{"", key}}
		tokens = append(tokens, t, xml.CharData(value), xml.EndElement{t.Name})
	}

	tokens = append(tokens, xml.EndElement{start.Name})

	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}

	// flush to ensure tokens are written
	err := e.Flush()
	if err != nil {
		return err
	}
	return nil
	*/
}

func (p *WxParams) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*p = WxParams{}
	for {
		var e xmlMapEntry

		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		(*p)[e.XMLName.Local] = e.Value
	}
	return nil
}

func (p WxParams)SortGenStr(seq string) string  {
	var keys []string
	var result string
	for k := range p {
		keys = append(keys,k)
	}
	sort.Strings(keys)
	for _,k := range keys {
		pv := ""
		switch p[k].(type) {
		case string:
			pv = p[k].(string)
		case int:
			pv = strconv.Itoa(p[k].(int))
		case WxParams:
			v,_ := json.Marshal(p[k])
			pv = string(v)
		}
		if len(pv) > 0 {
			if result == "" {
				result = k + "=" + pv
			} else {
				result += seq + k + "=" + pv
			}
		}
	}
	return result
}

func (p WxParams)GenStr(seq string) string  {
	buf := bytes.Buffer{}
	for k,v := range p {
		if buf.Len() > 0 {
			buf.WriteString(seq)
		}
		buf.WriteString(k)
		buf.WriteString("=")
		pv := ""
		switch v.(type) {
		case string:
			pv = v.(string)
		case int:
			pv = strconv.Itoa(p[k].(int))
		case WxParams:
			v,_ := json.Marshal(p[k])
			pv = string(v)
		}
		buf.WriteString(pv)
	}
	return buf.String()
}


func (p *WxParams)SetStr(s string,sep string)  {
	*p = make(WxParams)
	arr := strings.Split(s,sep)
	for _,value := range arr{
		items := strings.Split(value,"=")
		if len(items) == 2 {
			(*p)[items[0]] = items[1]
		}
	}
}

func (p *WxParams)SetUrlStr(s string)  {
	*p = make(WxParams)
	s2,_ := url.QueryUnescape(s)
	arr := strings.Split(s2,"&")
	for _,value := range arr{
		items := strings.Split(value,"=")
		if len(items) == 2 {
			(*p)[items[0]] = items[1]
		}
	}
}

func (p WxParams)GenUrlStr() string  {
	data := &url.Values{}
	for k, v := range p {
		if _,ok := v.(string);ok {
			data.Set(k, v.(string))
		} else {
			jsonv,_ := json.Marshal(v)
			data.Set(k, string(jsonv))
		}
	}
	return data.Encode()
}

func (p WxParams) MakeWxSign(signtype,apikey string) string {
	var signstr string
	s := p.SortGenStr("&")
	if len(s) > 0 {
		s += "&key=" + apikey
	}
	switch signtype {
	case Md5Sign:
		md := md5.Sum([]byte(s))
		signstr = hex.EncodeToString(md[:])
	case HMacSha256:
		h := hmac.New(sha256.New,[]byte(apikey))
		h.Write([]byte(s))
		signstr = hex.EncodeToString(h.Sum(nil))
	}
	return strings.ToUpper(signstr)
}

func (p WxParams)VerifyWxSign(signtype,apikey string) bool  {
	if _,has := p[WxSign]; !has {
		return false
	}
	sign := p[WxSign]
	tp := make(WxParams)
	for k,v := range p {
		tp[k] = v
	}
	delete(tp,WxSign)
	return tp.MakeWxSign(signtype,apikey) == sign
}

func (p WxParams)AppendSubParams(key string,ps AliParams) AliParams  {
	allps := make(AliParams)
	for k,v := range p {
		allps[k] = v
	}
	allps[key] = ps
	return allps
}

func (p WxParams) MakeAliSign(privkeybase64 string) string {
	s := p.SortGenStr("&")
	log.Println("sign str:",s)
	privkey,err := base64.StdEncoding.DecodeString(privkeybase64)
	if err != nil {
		return ""
	}
	rsa := util.NewrsaClient()
	rsa.SetPrivKey(privkey)
	signbyte,err := rsa.Sign([]byte(s))
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(signbyte)
}

func (p WxParams)VerifySyncAliSign(pubkeybase64 string) bool  {
	if _,has := p[AliSign]; !has {
		return false
	}
	sign := p[AliSign]
	tp := make(WxParams)
	for k,v := range p {
		tp[k] = v
	}
	delete(tp,WxSign)
	jsonstr,err := json.Marshal(tp)
	if err != nil {
		return false
	}
	pubkey,_ := base64.StdEncoding.DecodeString(pubkeybase64)
	signbyte,_ := base64.StdEncoding.DecodeString(sign.(string))
	rsa := util.NewrsaClient()
	rsa.SetPubKey(pubkey)
	return rsa.Verify([]byte(jsonstr),signbyte) == nil
}

func (p WxParams)VerifyAsyncAliSign(pubkeybase64 string) bool  {
	if _,has := p[AliSign]; !has {
		return false
	}
	sign := p[AliSign]
	tp := make(WxParams)
	for k,v := range p {
		if k != WxSign && k != SignType {
			tp[k] = v
		}
	}
	str := tp.SortGenStr("&")
	pubkey,_ := base64.StdEncoding.DecodeString(pubkeybase64)
	signbyte,_ := base64.StdEncoding.DecodeString(sign.(string))
	rsa := util.NewrsaClient()
	rsa.SetPubKey(pubkey)
	return rsa.Verify([]byte(str),signbyte) == nil
}

func MergeParams(ps ...WxParams) WxParams  {
	allps := make(WxParams)
	for _,p := range ps {
		if len(p) > 0 {
			for k,v := range p {
				allps[k] = v
			}
		}
	}
	return allps
}

func Buildaliform(url string,p map[string]string) string  {
	buf := bytes.Buffer{}
	buf.WriteString(`<form name="punchout_form" method="post" action="`)
	buf.WriteString(url)
	buf.WriteString(`">`)
	buf.WriteString("\n")
	for k,v := range p {
		buf.WriteString(`<input type="hidden" name="`)
		buf.WriteString(k)
		buf.WriteString(`" value="`)
		buf.WriteString(strings.Replace(v,`"`,"&quot;",-1))
		buf.WriteString(`">`)
		buf.WriteString("\n")
	}
	buf.WriteString(`<input type="submit" value="立即支付" style="display:none" >`)
	buf.WriteString("\n")
	buf.WriteString("</form>\n")
	buf.WriteString("<script>document.forms[0].submit();</script>")
	return buf.String()
}

func BuildwxjsCall(appid,prepay_id,returnurl,apikey string) string  {
	buf := bytes.Buffer{}
	params := make(WxParams)
	params["appId"] = appid
	params["timeStamp"] = util.CurTimeStamp()
	params["nonceStr"] = util.NonceStr(32)
	params["package"] = "prepay_id=" + prepay_id
	params["signType"] = Md5Sign
	sign := params.MakeWxSign(Md5Sign,apikey)
	params["paySign"] = sign
	jp,_ := json.Marshal(params)

	buf.WriteString("<script>\n")
	jstemplate := `function onBridgeReady(){
   WeixinJSBridge.invoke(
      'getBrandWCPayRequest',%s,
      function(res){
      if(res.err_msg == "get_brand_wcpay_request:ok" ){
		window.location = "%s";
      }else{
		alert(JSON.stringify(res));
      }
   });
}
if (typeof WeixinJSBridge == "undefined"){
   if( document.addEventListener ){
       document.addEventListener('WeixinJSBridgeReady', onBridgeReady, false);
   }else if (document.attachEvent){
       document.attachEvent('WeixinJSBridgeReady', onBridgeReady);
       document.attachEvent('onWeixinJSBridgeReady', onBridgeReady);
   }
}else{
   onBridgeReady();
}`
	buf.WriteString(fmt.Sprintf(jstemplate,jp,returnurl))
	buf.WriteString("\n</script>")

	/*
	buf.WriteString("<script>\n")
	buf.WriteString(`function onBridgeReady(){WeixinJSBridge.invoke('getBrandWCPayRequest', `)
	buf.WriteString(string(jp))
	buf.WriteString(`,function(res){`)
	if len(returnurl) > 0 {
		buf.WriteString(`if(res.err_msg == "get_brand_wcpay_request:ok" ){`)
		buf.WriteString("window.location = " + returnurl + ";}")
	}
	buf.WriteString("});}")
	buf.WriteString(`if (typeof WeixinJSBridge == "undefined"){
   if( document.addEventListener ){
       document.addEventListener('WeixinJSBridgeReady', onBridgeReady, false);
   }else if (document.attachEvent){
       document.attachEvent('WeixinJSBridgeReady', onBridgeReady); 
       document.attachEvent('onWeixinJSBridgeReady', onBridgeReady);
   }
}else{
   onBridgeReady();
}`)
	buf.WriteString("\n</script>")
*/
	return buf.String()
}

func BuildMsgPage(msg string) string  {
	return msg
}
