package httpd

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

func KfkCheck(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	//fmt.Println("kafka topic check ")
	url := "http://127.0.0.1/clusters/kafkalog_tj/consumers/newsclientLogConsumerES72/topic/newsclientLog/type/KF"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Loges.Error("error : ", zap.Error(err))
	}

	req.SetBasicAuth("admin", "passwd")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		Loges.Error("error : ", zap.Error(err))
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		Loges.Info("status code error:", zap.Any("statusCode", res.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		Loges.Error("error : ", zap.Error(err))
	}

	doc.Find("tr:contains(Total)").Each(func(i int, selection *goquery.Selection) {
		ss := selection.Text()

		aa := strings.SplitAfter(ss, "Total Lag")

		dd, err := strconv.Atoi(strings.ReplaceAll(strings.TrimSpace(aa[1]), ",", ""))
		if err != nil {
			Loges.Error("error : ", zap.Error(err))
		}
		if dd >= 10000000 {
			Loges.Info("kafka total lag is error :", zap.Any("total lag", dd))
			SendWechat("newsClientES total lag :", strings.ReplaceAll(strings.TrimSpace(aa[1]), ",", ""))
		}

	})
	w.Write([]byte("check ok"))

}
