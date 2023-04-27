package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/niubir/wechat"
)

var (
	w         *wechat.Wechat
	appid     string
	appsecret string
	token     string
	domain    string
	port      string

	oauth2Redirect string
)

func init() {
	appid = os.Getenv("APPID")
	appsecret = os.Getenv("APPSECRET")
	token = os.Getenv("TOKEN")
	domain = os.Getenv("DOMAIN")
	port = os.Getenv("PORT")

	if port == "" {
		port = "80"
	}

	oauth2Redirect = fmt.Sprintf("%s:%s/oauth2/callback", domain, port)

	w = wechat.NewWechat(appid, appsecret, token)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/configure", wechatConfigureHandler)
	http.HandleFunc("/oauth2", wechatOAuth2Handler)
	http.HandleFunc("/oauth2/callback", wechatOAuth2CallbackHandler)

	if err := http.ListenAndServe(":", nil); err != nil {
		panic(err)
	}
}

func indexHandler(resp http.ResponseWriter, req *http.Request) {
	success(resp, []byte("Wellcome!"))
}

func wechatConfigureHandler(resp http.ResponseWriter, req *http.Request) {
	echostr, err := w.Configure(req)
	if err != nil {
		failed(resp, err)
		return
	}
	success(resp, []byte(echostr))
}

func wechatOAuth2Handler(resp http.ResponseWriter, req *http.Request) {
	// url, err := w.GetOAuthURL(oauth2Redirect, "", true)
	url, err := w.GetOAuthURL(oauth2Redirect, "This is test state")
	if err != nil {
		failed(resp, err)
		return
	}
	redirect(resp, req, url)
}

func wechatOAuth2CallbackHandler(resp http.ResponseWriter, req *http.Request) {
	code, err := w.ParseOAuthCode(req)
	if err != nil {
		failed(resp, err)
		return
	}
	accessToken, err := w.GetOAuthAccessToken(code.Code)
	if err != nil {
		failed(resp, err)
		return
	}

	data := map[string]interface{}{
		"accessToken": accessToken,
	}

	if accessToken.ContainScopeUserinfo() {
		user, err := w.GetOAuthUserByAccessToken(*accessToken)
		if err != nil {
			failed(resp, err)
			return
		}
		data["user"] = *user
	}

	successJSON(resp, data)
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
