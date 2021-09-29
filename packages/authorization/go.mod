module noz.zkip.cc/auth

go 1.17

require (
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.3 // indirect
	noz.zkip.cc/utils v0.0.0
)

require (
	github.com/go-redis/redis v6.15.9+incompatible // indirect
	github.com/twinj/uuid v1.0.0 // indirect
)

replace noz.zkip.cc/utils => ../utils
