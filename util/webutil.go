package util

import (
	"crypto/tls"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func NewHttpsClient(cert *tls.Certificate) *http.Client  {
	config := &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	transport := &http.Transport{
		TLSClientConfig:    config,
		DisableCompression: true,
	}
	h := &http.Client{Transport: transport}
	return h
}

// 用时间戳生成随机字符串
func CurTimeStamp() string {
	//return strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	return strconv.FormatInt(time.Now().Unix(), 10)
}

//生成随机字符串
func NonceStr(length int) string{
	var r []byte
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		r = append(r,str[rand.Intn(len(str))])
	}
	return string(r)
}

func BuildUrlCode(mp map[string]string) string  {
	data := &url.Values{}
	for k, v := range mp {
		data.Set(k,v)
	}
	return data.Encode()
}