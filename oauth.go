package wechat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	oauth2_url              = "https://open.weixin.qq.com/connect/oauth2/authorize?%s#wechat_redirect"
	oauth2_access_token_url = "https://api.weixin.qq.com/sns/oauth2/access_token?%s"
	oauth2_user_info_url    = "https://api.weixin.qq.com/sns/userinfo?%s"
)

var (
	oauth_state_reg = regexp.MustCompile(`^[A-Za-z0-9]{0,128}$`)
)

type (
	OAuthCode struct {
		Code  string
		State string
	}
	OAuthAccessToken struct {
		AccessToken    string `json:"access_token"`
		RefreshToken   string `json:"refresh_token"`
		ExpiresIn      int64  `json:"expires_in"`
		Openid         string `json:"openid"`
		Unionid        string `json:"unionid"`
		Scope          string `json:"scope"`
		IsSnapshotuser int    `json:"is_snapshotuser"`
	}
	OAuthUser struct {
		Openid     string   `json:"openid"`
		Unionid    string   `json:"unionid"`
		Nickname   string   `json:"nickname"`
		Headimgurl string   `json:"headimgurl"`
		Sex        int      `json:"sex"`
		Province   string   `json:"province"`
		City       string   `json:"city"`
		Country    string   `json:"country"`
		Privilege  []string `json:"privilege"`
	}
)

func (w *Wechat) GetOAuthURL(redirect, state string, isScopeBase ...bool) (string, error) {
	if !oauth_state_reg.MatchString(state) {
		return "", errors.New("invalid state")
	}

	v := url.Values{}
	v.Add("appid", w.appid)
	v.Add("redirect_uri", redirect)
	v.Add("response_type", "code")
	if len(isScopeBase) > 0 && isScopeBase[0] {
		v.Add("scope", "snsapi_base")
	} else {
		v.Add("scope", "snsapi_userinfo")
	}
	v.Add("state", state)
	return fmt.Sprintf(oauth2_url, v.Encode()), nil
}

func (w *Wechat) ParseOAuthCode(req *http.Request) (*OAuthCode, error) {
	code := req.URL.Query().Get("code")
	state := req.URL.Query().Get("state")
	if code == "" {
		return nil, errors.New("invalid code")
	}
	return &OAuthCode{
		Code:  code,
		State: state,
	}, nil
}

func (w *Wechat) GetOAuthAccessToken(code string) (*OAuthAccessToken, error) {
	if code == "" {
		return nil, errors.New("invalid code")
	}

	v := url.Values{}
	v.Add("appid", w.appid)
	v.Add("secret", w.appsecret)
	v.Add("code", code)
	v.Add("grant_type", "authorization_code")

	resp, err := http.Get(fmt.Sprintf(oauth2_access_token_url, v.Encode()))
	if err != nil {
		return nil, err
	}

	var accessToken OAuthAccessToken
	if err := w.exchangeOAuthResponse(resp, &accessToken); err != nil {
		return nil, err
	}

	return &accessToken, nil
}

func (w *Wechat) GetOAuthUserByCode(code string) (*OAuthUser, error) {
	accessToken, err := w.GetOAuthAccessToken(code)
	if err != nil {
		return nil, err
	}
	return w.GetOAuthUserByAccessToken(*accessToken)
}

func (w *Wechat) GetOAuthUserByAccessToken(accessToken OAuthAccessToken) (*OAuthUser, error) {
	if accessToken.AccessToken == "" {
		return nil, errors.New("invalid access token")
	}
	if accessToken.Openid == "" {
		return nil, errors.New("invalid openid")
	}
	if !accessToken.ContainScopeUserinfo() {
		return nil, errors.New("limit by scope_userinfo")
	}

	v := url.Values{}
	v.Add("access_token", accessToken.AccessToken)
	v.Add("openid", accessToken.Openid)
	v.Add("lang", "zh_CN")

	resp, err := http.Get(fmt.Sprintf(oauth2_user_info_url, v.Encode()))
	if err != nil {
		return nil, err
	}

	var user OAuthUser
	if err := w.exchangeOAuthResponse(resp, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (w *Wechat) exchangeOAuthResponse(resp *http.Response, data interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return w.exchangeOAuthBody(body, data)
}

func (w *Wechat) exchangeOAuthBody(body []byte, data interface{}) error {
	var codemsg struct {
		Errcode string `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &codemsg); err != nil {
		return err
	}
	if codemsg.Errcode != "" {
		return fmt.Errorf("%s: %s", codemsg.Errcode, codemsg.Errmsg)
	}
	return json.Unmarshal(body, data)
}

func (accessToken OAuthAccessToken) ContainScopeBase() bool {
	return strings.Contains(accessToken.Scope, "snsapi_base")
}

func (accessToken OAuthAccessToken) ContainScopeUserinfo() bool {
	return strings.Contains(accessToken.Scope, "snsapi_userinfo")
}
