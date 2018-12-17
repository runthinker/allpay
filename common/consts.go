package common

const (
	Fail                       = "FAIL"
	Success                    = "SUCCESS"
	BodyXMLType 			   = "application/xml; charset=utf-8"
	BodyJSONType 			   = "application/json; charset=utf-8"
	BodyUrlencodedType 		   = "application/x-www-form-urlencoded; charset=utf-8"
	ALISUCCESS				   = "10000"

	MicroPayUrl                = "https://api.mch.weixin.qq.com/pay/micropay"
	UnifiedOrderUrl            = "https://api.mch.weixin.qq.com/pay/unifiedorder"
	OrderQueryUrl              = "https://api.mch.weixin.qq.com/pay/orderquery"
	ReverseUrl                 = "https://api.mch.weixin.qq.com/secapi/pay/reverse"
	CloseOrderUrl              = "https://api.mch.weixin.qq.com/pay/closeorder"
	RefundUrl                  = "https://api.mch.weixin.qq.com/secapi/pay/refund"
	RefundQueryUrl             = "https://api.mch.weixin.qq.com/pay/refundquery"
	DownloadBillUrl            = "https://api.mch.weixin.qq.com/pay/downloadbill"
	ReportUrl                  = "https://api.mch.weixin.qq.com/payitil/report"
	ShortUrl                   = "https://api.mch.weixin.qq.com/tools/shorturl"
	AuthCodeToOpenidUrl        = "https://api.mch.weixin.qq.com/tools/authcodetoopenid"

	SandboxMicroPayUrl         = "https://api.mch.weixin.qq.com/sandboxnew/pay/micropay"
	SandboxUnifiedOrderUrl     = "https://api.mch.weixin.qq.com/sandboxnew/pay/unifiedorder"
	SandboxOrderQueryUrl       = "https://api.mch.weixin.qq.com/sandboxnew/pay/orderquery"
	SandboxReverseUrl          = "https://api.mch.weixin.qq.com/sandboxnew/secapi/pay/reverse"
	SandboxCloseOrderUrl       = "https://api.mch.weixin.qq.com/sandboxnew/pay/closeorder"
	SandboxRefundUrl           = "https://api.mch.weixin.qq.com/sandboxnew/secapi/pay/refund"
	SandboxRefundQueryUrl      = "https://api.mch.weixin.qq.com/sandboxnew/pay/refundquery"
	SandboxDownloadBillUrl     = "https://api.mch.weixin.qq.com/sandboxnew/pay/downloadbill"
	SandboxReportUrl           = "https://api.mch.weixin.qq.com/sandboxnew/payitil/report"
	SandboxShortUrl            = "https://api.mch.weixin.qq.com/sandboxnew/tools/shorturl"
	SandboxAuthCodeToOpenidUrl = "https://api.mch.weixin.qq.com/sandboxnew/tools/authcodetoopenid"
)

const (
	K_ALI_PAY_TRADE_STATUS_WAIT_BUYER_PAY = "WAIT_BUYER_PAY" // 交易创建，等待买家付款
	K_ALI_PAY_TRADE_STATUS_TRADE_CLOSED   = "TRADE_CLOSED"   // 未付款交易超时关闭，或支付完成后全额退款
	K_ALI_PAY_TRADE_STATUS_TRADE_SUCCESS  = "TRADE_SUCCESS"  // 交易支付成功
	K_ALI_PAY_TRADE_STATUS_TRADE_FINISHED = "TRADE_FINISHED" // 交易结束，不可退款

	K_ALI_PAY_SANDBOX_API_URL     = "https://openapi.alipaydev.com/gateway.do"
	K_ALI_PAY_PRODUCTION_API_URL  = "https://openapi.alipay.com/gateway.do"
	K_ALI_PAY_PRODUCTION_MAPI_URL = "https://mapi.alipay.com/gateway.do"
)