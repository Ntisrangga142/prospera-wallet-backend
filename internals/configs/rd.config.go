package configs

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	rdbHost := os.Getenv("RDBHOST")
	rdbPort := os.Getenv("RDBPORT")
	rdbUser := os.Getenv("RDBUSER")
	rdbPass := os.Getenv("RDBPASS")
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", rdbHost, rdbPort),
		Username: rdbUser,
		Password: rdbPass,
	})
}
