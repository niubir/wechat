package wechat

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/http"
	"sort"
	"strings"
)

func (w *Wechat) Configure(req *http.Request) (string, error) {
	query := req.URL.Query()
	signature := query.Get("signature")
	timestamp := query.Get("timestamp")
	nonce := query.Get("nonce")
	echostr := query.Get("echostr")

	if err := w.checkConfigureSign(signature, nonce, timestamp, w.token); err != nil {
		return "", err
	}
	return echostr, nil
}

func (w *Wechat) checkConfigureSign(signature string, s ...string) error {
	if w.configureSign(s...) != signature {
		return errors.New("invalid signature")
	}
	return nil
}

func (w *Wechat) configureSign(s ...string) string {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
	o := sha1.New()
	o.Write([]byte(strings.Join(s, "")))
	return hex.EncodeToString(o.Sum(nil))
}
