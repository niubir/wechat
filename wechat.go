package wechat

type Wechat struct {
	appid     string
	appsecret string
	token     string
}

func NewWechat(appid, appsecret, token string) *Wechat {
	return &Wechat{
		appid:     appid,
		appsecret: appsecret,
		token:     token,
	}
}
