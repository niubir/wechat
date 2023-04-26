package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/niubir/wechat"
)

var (
	w              *wechat.Wechat
	appid          = os.Getenv("APPID")
	appsecret      = os.Getenv("APPSECRET")
	token          = os.Getenv("TOKEN")
	oauth2Redirect = os.Getenv("OAUTHDOMAIN") + "/oauth2/callback"
)

func init() {
	w = wechat.NewWechat(appid, appsecret, token)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/configure", wechatConfigureHandler)
	http.HandleFunc("/oauth2", wechatOAuth2Handler)
	http.HandleFunc("/oauth2/callback", wechatOAuth2CallbackHandler)

	if err := http.ListenAndServe(":10001", nil); err != nil {
		panic(err)
	}
}

func indexHandler(resp http.ResponseWriter, req *http.Request) {
	success(resp, []byte("Wellcome 10001!"))
}

func wechatConfigureHandler(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("wechatConfigureHandler1")
	echostr, err := w.Configure(req)
	if err != nil {
		failed(resp, err)
		return
	}
	success(resp, []byte(echostr))
}

func wechatOAuth2Handler(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("wechatOAuth2Handler1")
	url, err := w.OAuth2(oauth2Redirect, wechat.ScopeUserInfo, "")
	// url, err := w.OAuth2(oauth2Redirect, wechat.ScopeBase, "")
	if err != nil {
		failed(resp, err)
		return
	}
	fmt.Printf("url: %s\n", url)
	successJSON(resp, url)
	// redirect(resp, req, url)
}

func wechatOAuth2CallbackHandler(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("wechatOAuth2CallbackHandler1")
	code, err := w.GetOAuth2Code(req)
	if err != nil {
		failed(resp, err)
		return
	}
	fmt.Printf("code: %+v\n", code)
	accessToken, err := w.GetOAuth2AccessToken(code.Code)
	if err != nil {
		failed(resp, err)
		return
	}
	fmt.Printf("accessToken: %+v\n", accessToken)
	user, err := w.GetOAuth2User(*accessToken, wechat.LangZH_TW)
	if err != nil {
		failed(resp, err)
		return
	}
	fmt.Printf("user: %+v\n", user)
	successJSON(resp, user)
}

func failed(resp http.ResponseWriter, err error) {
	resp.WriteHeader(http.StatusBadRequest)
	resp.Write([]byte("FAILED:" + err.Error()))
}

func successJSON(resp http.ResponseWriter, data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	resp.Write(body)
}

func success(resp http.ResponseWriter, body []byte) {
	resp.Write(body)
}

func redirect(resp http.ResponseWriter, req *http.Request, url string) {
	resp.WriteHeader(http.StatusFound)
	http.Redirect(resp, req, url, http.StatusFound)
}
