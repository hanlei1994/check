package httpd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/spf13/viper"
)

type HttpBody struct {
	code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

func GetJwt() string {
	jwtUrl := viper.GetString("SendMsg.MsgGw") + "/api/v1/jwt"
	req, err := http.NewRequest("GET", jwtUrl, nil)
	if err != nil {
		Loges.Error("new request is err:", zap.Error(err))
	}
	req.Header.Add("AppID", viper.GetString("SendMsg.AppID"))
	req.Header.Add("AppKey", viper.GetString("SendMsg.AppKey"))

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		Loges.Error("do request is err:", zap.Error(err))
	}

	defer res.Body.Close()

	s, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Loges.Error("read res.body is err:", zap.Error(err))
	}
	jh := HttpBody{}
	err = json.Unmarshal(s, &jh)
	if err != nil {
		Loges.Error("json exchange is err:", zap.Error(err))
	}

	return jh.Data

}

func SendWechat(msg ...string) error {

	sendUrl := viper.GetString("SendMsg.MsgGw") + "/api/v1/wechat"
	msg = append(msg, time.Now().Format(time.RFC3339))
	data := url.Values{
		"account": {viper.GetString("SendMsg.To")},
		"title":   {msg[0]},
		"content": {strings.Join(msg, "\n")},
	}

	req, err := http.NewRequest("POST", sendUrl, strings.NewReader(data.Encode()))
	if err != nil {
		Loges.Error("new request is err:", zap.Error(err))
	}
	ss := GetJwt()
	req.Header.Add("AppJWTKey", ss)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		Loges.Error("do request is err:", zap.Error(err))
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return err
	}

	s, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Loges.Error("read res.body  is err:", zap.Error(err))
	}

	Loges.Info("send msg code", zap.ByteString("info", s))

	return nil
}
