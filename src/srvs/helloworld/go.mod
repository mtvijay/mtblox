module srvs/helloworld

go 1.15

require (
        github.com/mtbox/metrics v0.0.0
        github.com/mtbox/mtlog v0.0.0
        github.com/mtbox/mtsrv v0.0.0
)

replace github.com/mtbox/metrics => ./../../libs/metrics
replace github.com/mtbox/mtlog => ./../../libs/mtlog
replace github.com/mtbox/mtsrv => ./../../libs/mtsrv
