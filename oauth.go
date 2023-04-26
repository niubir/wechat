package wechat

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	ScopeBase     = "snsapi_base"
	ScopeUserInfo = "snsapi_userinfo"

	LangZH_CN = "zh_CN"
	LangZH_TW = "zh_TW"
	LangEN    = "en"

	oauth2_template              = "https://open.weixin.qq.com/connect/oauth2/authorize?%s#wechat_redirect"
	oauth2_access_token_template = "https://api.weixin.qq.com/sns/oauth2/access_token?%s"
	oauth2_user_info_template    = "https://api.weixin.qq.com/sns/userinfo?%s"
)

var (
	oauth_state_reg = regexp.MustCompile(`^[A-Za-z0-9]{0,128}$`)
)

type OAuth2Code struct {
	Code  string
	State string
}

type OAuth2AccessToken struct {
	AccessToken    string `json:"access_token"`
	ExpiresIn      int64  `json:"expires_in"`
	RefreshToken   string `json:"refresh_token"`
	Openid         string `json:"openid"`
	Scope          string `json:"scope"`
	IsSnapshotuser int    `json:"is_snapshotuser"`
	Unionid        string `json:"unionid"`
}

type OAuth2User struct {
	Openid     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	Headimgurl string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	Unionid    string   `json:"unionid"`
}

func (w *Wechat) OAuth2(redirect, scope, state string) (string, error) {
	if scope != ScopeBase && scope != ScopeUserInfo {
		return "", errors.New("invalid scope")
	}
	if !oauth_state_reg.MatchString(state) {
		return "", errors.New("invalid state")
	}

	v := url.Values{}
	v.Add("appid", w.appid)
	v.Add("redirect_uri", redirect)
	v.Add("response_type", "code")
	v.Add("scope", scope)
	v.Add("state", state)
	return fmt.Sprintf(oauth2_template, v.Encode()), nil
}

func (w *Wechat) GetOAuth2Code(req *http.Request) (*OAuth2Code, error) {
	query := req.URL.Query()
	return &OAuth2Code{
		Code:  query.Get("code"),
		State: query.Get("state"),
	}, nil
}

func (w *Wechat) GetOAuth2AccessToken(code string) (*OAuth2AccessToken, error) {
	if code == "" {
		return nil, errors.New("invalid scope")
	}
	v := url.Values{}
	v.Add("appid", w.appid)
	v.Add("secret", w.appsecret)
	v.Add("code", code)
	v.Add("grant_type", "authorization_code")

	resp, err := http.Get(fmt.Sprintf(oauth2_access_token_template, v.Encode()))
	if err != nil {
		return nil, err
	}
	var ac OAuth2AccessToken
	if err := w.parseBody(resp, &ac); err != nil {
		return nil, err
	}
	return &ac, nil
}

func (w *Wechat) GetOAuth2User(accessToken OAuth2AccessToken, lang string) (*OAuth2User, error) {
	if accessToken.AccessToken == "" {
		return nil, errors.New("invalid access token")
	}
	if accessToken.Openid == "" {
		return nil, errors.New("invalid openid")
	}
	if !strings.Contains(accessToken.Scope, ScopeUserInfo) {
		return nil, errors.New("invalid scope")
	}
	if lang != LangZH_CN && lang != LangZH_TW && lang != LangEN {
		return nil, errors.New("invalid lang")
	}
	v := url.Values{}
	v.Add("access_token", accessToken.AccessToken)
	v.Add("openid", accessToken.Openid)
	v.Add("lang", lang)

	resp, err := http.Get(fmt.Sprintf(oauth2_user_info_template, v.Encode()))
	if err != nil {
		return nil, err
	}
	var au OAuth2User
	if err := w.parseBody(resp, &au); err != nil {
		return nil, err
	}
	return &au, nil
}
