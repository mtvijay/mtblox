module libs/mtsrv

go 1.15

require (
        github.com/mtbox/metrics v0.0.0
        github.com/mtbox/mtlog v0.0.0

	github.com/shirou/gopsutil v3.20.12+incompatible
	google.golang.org/grpc v1.34.0
)

replace github.com/mtbox/metrics => ./../metrics
replace github.com/mtbox/mtlog => ./../mtlog
