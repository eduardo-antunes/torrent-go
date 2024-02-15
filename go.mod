module github.com/eduardo-antunes/torrent-go

go 1.21.7

require internal/metainfo v1.0.0

require github.com/mitchellh/mapstructure v1.5.0 // indirect

replace internal/metainfo => ./internal/metainfo
