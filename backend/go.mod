module github.com/yaanno/backend

go 1.21

require (
	github.com/google/uuid v1.6.0
	github.com/nsqio/go-nsq v1.1.0
	github.com/prometheus/client_golang v1.20.5
	github.com/rs/zerolog v1.31.0
	github.com/sony/gobreaker v1.0.0
	github.com/yaanno/shared v0.0.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	golang.org/x/sys v0.22.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace github.com/yaanno/shared => ../shared
