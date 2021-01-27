module srvs/helloworld

go 1.15

require (
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.2
	github.com/gorilla/csrf v1.7.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/mtbox/metrics v0.0.0
	github.com/mtbox/mtlog v0.0.0
	github.com/mtbox/mtsrv v0.0.0
)

replace github.com/mtbox/metrics => ./../../libs/metrics

replace github.com/mtbox/mtlog => ./../../libs/mtlog

replace github.com/mtbox/mtsrv => ./../../libs/mtsrv
