package model

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"os"
	"encoding/json"
	"log"
	"github.com/runthinker/allpay/wxpay"
)

type (
	WxConfig struct {
		Apikey string `json:"apikey"`
		Appid string `json:"appid"`
		MpSecret string `json:"mp_secret"`
		MchId string `json:"mch_id"`
		Certfilepath string `json:"certfilepath"`
		NotifyUrl string `json:"notify_url"`
		SandBoxSignkey string `json:"sand_box_signkey"`
	}
	AliConfig struct {
		MchPrivatekey string `json:"mch_privatekey"`
		AliPublickey string `json:"ali_publickey"`
		AppId string `json:"app_id"`
		ProviderId string `json:"provider_id"`
		NotifyUrl string `json:"notify_url"`
	}
	DbConfig struct {
		Dbpath string `json:"dbpath"`
		Sqldebug bool `json:"sqldebug"`
	}

	AllPayConfig struct {
		WxPay WxConfig
		AliPay AliConfig
		Db	DbConfig
	}
)

var gdb *gorm.DB
var ApConfig *AllPayConfig

func init()  {
	ApConfig = &AllPayConfig{
		WxConfig{
			Apikey:       "",
			Appid:        "",
			MpSecret:	  "",
			MchId:        "",
			Certfilepath: "",
			NotifyUrl:    "http://test.xxxx.com/payapi/notify/wxpay",
		},
		AliConfig{
			MchPrivatekey: ``,
			AliPublickey:  ``,
			AppId:         "",
			ProviderId:    "",
			NotifyUrl:     "http://test.xxxx.com/payapi/notify/alipay",
		},
		DbConfig{"root:123456@/allpay?charset=utf8&parseTime=True&loc=Local",true},
	}
	file,err := os.Open("config.json")
	defer file.Close()
	if err == nil {
		err = json.NewDecoder(file).Decode(&ApConfig)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}
	if signkey := wxpay.GetSandBoxSignKey(ApConfig.WxPay.MchId,ApConfig.WxPay.Apikey); signkey != "" {
		ApConfig.WxPay.SandBoxSignkey = signkey
		log.Println("SandboxSignkey:", ApConfig.WxPay.SandBoxSignkey)
	}
	gdb, err = gorm.Open("mysql",ApConfig.Db.Dbpath)
	if err != nil {
		log.Fatalf("Fail to create engine: %v\n", err)
	}
	if ApConfig.Db.Sqldebug {
		gdb.LogMode(true)
	}
	gdb.SingularTable(true)
	gdb.DB().SetMaxIdleConns(20)
	gdb.DB().SetMaxIdleConns(2)
	if err := gdb.DB().Ping();err != nil {
		log.Fatalln(err)
	}
}
