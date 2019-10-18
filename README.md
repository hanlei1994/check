# this is check service 
## env
```
go version 1.12.5
linux
```
## api info 
```$xslt
http://127.0.0.1:7789/checkport   #check service port,tcp only
http://127.0.0.1:7789/checkpod    #check k8s pods running status
http://127.0.0.1:7789/checkmgo  #check mongo running status
http://127.0.0.1:7789/checkes     #check es-cluster running status
http://127.0.0.1:7789/checknode     #check k8scluster nodes running status,,,error
http://127.0.0.1:7789/checktotalag   #check kafka consumer total lag 

```
http://127.0.0.1:7789/postport   #post port data to db
eg:
```cassandraql
{"topic":"test","data":"127.0.0.1:1234"}
```



