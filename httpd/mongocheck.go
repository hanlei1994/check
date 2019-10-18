package httpd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/julienschmidt/httprouter"

	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MemStat struct {
	Name   string `json:"name"`
	Health int    `json:"health"`
	Self   bool   `json:"self"`
}

type MgodStats struct {
	Set     string    `json:"set"`
	Members []MemStat `json:"members"`
	Ok      int       `json:"ok"`
}

type HttpMgo struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data []MgodStats `json:"data"`
}

func MongoCheck(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, GET, OPTIONS, POST, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Accept-Encoding, Authorization, Content-Length, Content-Type, X-CSRF-Token, X-HTTP-Method-Override, AuthToken, NS, Resource, X-Requested-With")

	htpSta := HttpMgo{}

	for _, conn := range viper.GetStringSlice("mongo.conn") {

		ch := make(chan MgodStats)

		go func(conn string) {

			sess, err := mgo.Dial(conn)
			if err != nil {
				fmt.Println("11111111", err)
				ch <- MgodStats{Ok: 0, Set: conn}
				return

			}
			defer sess.Close()

			err = sess.DB(viper.GetString("mongo.replset_authdb")).Login(viper.GetString("mongo.replset_user"), viper.GetString("mongo.replset_passwd"))
			if err != nil {
				Loges.Error("auth mgo is err:", zap.Error(err))
				ch <- MgodStats{Ok: 0, Set: conn}
				return
			}

			result := MgodStats{}

			err = sess.DB("admin").Run(bson.D{{"replSetGetStatus", 1}}, &result)
			if err != nil {
				Loges.Error("mgo run is err:", zap.Error(err))
			}
			var i int
			for _, v := range result.Members {
				if v.Health == 1 {
					i = i + 1
				}
			}

			if i != len(result.Members) {
				ch <- MgodStats{Ok: 0, Members: result.Members, Set: result.Set}
			} else {
				ch <- MgodStats{Ok: result.Ok, Members: result.Members, Set: result.Set}
			}

		}(conn)

		v := <-ch

		htpSta.Data = append(htpSta.Data, v)

	}
	htpSta.Code = 200
	htpSta.Msg = "green"
	for _, k := range htpSta.Data {
		if k.Ok != 1 {
			htpSta.Msg = "yellow"
			for _, j := range k.Members {
				if j.Health != 1 {
					SendWechat(j.Name, GetHostNname(j.Name))
				}
			}
		}
	}

	w2, _ := json.Marshal(htpSta)
	w.Write(w2)

}
