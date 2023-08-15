package redis

const script = `
local current
current = redis.call("INCR", KEYS[1])
if tonumber(current) == 1 then
	redis.call("EXPIRE", KEYS[1], 60)
end
return current
`

// https://hjr265.me/blog/simple-rate-limiter-with-redis/
// Note that a fixed-window rate limiter, although effective against sustained attacks, may affect the experience of legitimate users.
//
// We rate limit based on the username used during a login flow. This is less likely to affect legitimate users than using, for example, the remote IP address of the incoming request.
// func isRateLimited(ctx context.Context, key string, limit int64) (bool, error) {
// 	v, err := redisClient.Eval(ctx, script, []string{key}).Result()
// 	if err != nil {
// 		return false, err
// 	}
// 	n, _ := v.(int64)
// 	return n > int64(limit), nil
// }
