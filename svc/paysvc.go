package main

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"allpay/svc/api"
)

func VerifyUrlSign(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer func() {
			if e := recover(); e != nil {
				log.Println(r.URL.Path,"Panicing error:", e)
			}
		}()
		/*
		暂时不验签
		url := r.URL.Query()
		params := make(common.WxParams)
		for k,v := range url {
			params[k] = v[0]
		}
		*/
		next(w,r,p)
	}
}

func main() {
	h := httprouter.New()
	h.GET("/payapi/test", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.Write([]byte("server is ready!"))
	})
	h.GET("/payapi/wappay/:mode",VerifyUrlSign(api.Wap.DoPagePay))
	h.POST("/payapi/notify/:paytype",api.Wap.NotifyPay)
	http.ListenAndServe("localhost:3003",h)
}
