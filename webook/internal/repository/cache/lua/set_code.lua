local key = KEYS[1]
local cntKey = key..":cnt"

local val = ARGV[1]

local ttl = tonumber(redis.call("ttl", key))
-- -1 是存在但没有过期时间
if ttl == -1 then
    return -2

-- -2 是key不存在
elseif ttl == -2 or ttl < 540 then
    redis.call("set", key, val)
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call(expire, cntKey, 600)
    return 0
else
    -- 已经发送了一个验证码，还不到1分钟
    return -1
end