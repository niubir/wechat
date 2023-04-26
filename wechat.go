package wechat

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

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

func (w *Wechat) Configure(req *http.Request) (string, error) {
	query := req.URL.Query()
	signature := query.Get("signature")
	timestamp := query.Get("timestamp")
	nonce := query.Get("nonce")
	echostr := query.Get("echostr")

	if err := w.checkSign(signature, nonce, timestamp, w.token); err != nil {
		return "", err
	}
	return echostr, nil
}

func (w *Wechat) checkSign(signature string, s ...string) error {
	if w.sign(s...) != signature {
		return errors.New("invalid signature")
	}
	return nil
}

func (w *Wechat) sign(s ...string) string {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
	o := sha1.New()
	o.Write([]byte(strings.Join(s, "")))
	return hex.EncodeToString(o.Sum(nil))
}

func (w *Wechat) parseBody(resp *http.Response, data interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
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
