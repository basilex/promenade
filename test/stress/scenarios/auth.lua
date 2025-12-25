-- wrk Load Test: Authentication Flow
-- Tests login endpoint with realistic user credentials

-- Configuration
local counter = 0
local users = {
    {email = "admin@promenade.com", password = "passw0rd"},
    {email = "moderator@promenade.com", password = "passw0rd"},
    {email = "alexander.vasilenko@gmail.com", password = "03041965"}
}

-- Initialize
function init(args)
    math.randomseed(os.time())
end

-- Generate request
function request()
    counter = counter + 1
    local user = users[(counter % #users) + 1]
    
    local body = string.format([[
{
    "email": "%s",
    "password": "%s"
}]], user.email, user.password)
    
    local path = "/api/v1/auth/login"
    local headers = {
        ["Content-Type"] = "application/json"
    }
    
    return wrk.format("POST", path, headers, body)
end

-- Handle response
function response(status, headers, body)
    if status ~= 200 then
        print(string.format("Error: %d - %s", status, body))
    end
end

-- Summary statistics
function done(summary, latency, requests)
    io.write("------------------------------\n")
    io.write(string.format("  Total requests: %d\n", summary.requests))
    io.write(string.format("  Total duration: %.2fs\n", summary.duration / 1000000))
    io.write(string.format("  Requests/sec:   %.2f\n", summary.requests / (summary.duration / 1000000)))
    io.write(string.format("  Total errors:   %d\n", summary.errors.connect + summary.errors.read + summary.errors.write + summary.errors.timeout))
    io.write("------------------------------\n")
end
