package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Mongo struct {
		Uri    string
		DbName string
	}
	RedisConf struct {
		Host string
		Pass string
	}
}
