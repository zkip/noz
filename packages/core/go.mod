module noz.zkip.cc/core

go 1.17

require (
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/gabriel-vasile/mimetype v1.3.1 // indirect
	github.com/go-redis/redis v6.15.9+incompatible // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/twinj/uuid v1.0.0 // indirect
	golang.org/x/net v0.0.0-20210505024714-0287a6fb4125 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

require noz.zkip.cc/utils v0.0.0

replace noz.zkip.cc/utils => ../utils
