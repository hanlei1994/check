package main //import "check"

import (
	"check/httpd"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/julienschmidt/httprouter"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("conf")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		//fmt.Printf("Config file changed: %s", e.Name)
		httpd.Loges.Info("Config file changed: ", zap.Any("", e.Name))
	})

	ClusterCfgs := httpd.GetConfig()

	httpd.Loges.Info("this is log test")
	router := httprouter.New()
	router.GET("/checkport", httpd.PortCheck)
	router.GET("/checkpod", httpd.Podcheck(ClusterCfgs))
	router.GET("/checkmgo", httpd.MongoCheck)
	router.GET("/checkes", httpd.Escheck)
	router.GET("/checktotalag", httpd.KfkCheck)
	router.GET("/checknode", httpd.NodeCheck(ClusterCfgs))
	router.POST("/postport", httpd.PostPort())

	//log.Fatal(http.ListenAndServe(viper.GetString("check.hostPort"), router))
	httpd.Loges.Fatal("services :", zap.Any("", http.ListenAndServe(viper.GetString("check.hostPort"), router)))

}
