-- wrk Load Test: Realistic User Journey
-- Simulates real user behavior: login → browse posts → create post → logout

-- Configuration
local TOKEN = nil
local USER_STATE = {}
local counter = 0

-- User journey states
local STATES = {
    LOGIN = 1,
    BROWSE = 2,
    CREATE = 3,
    READ = 4
}

-- Initialize
function init(args)
    math.randomseed(os.time())
    USER_STATE = {
        state = STATES.LOGIN,
        post_id = nil,
        browse_count = 0
    }
end

-- Setup thread
function setup(thread)
    thread:set("id", counter)
    counter = counter + 1
end

-- Generate request
function request()
    local state = USER_STATE.state
    
    if state == STATES.LOGIN then
        -- User logs in
        USER_STATE.state = STATES.BROWSE
        
        local body = [[
{
    "email": "admin@promenade.com",
    "password": "passw0rd"
}]]
        
        return wrk.format("POST", "/api/v1/auth/login", {
            ["Content-Type"] = "application/json"
        }, body)
        
    elseif state == STATES.BROWSE then
        -- User browses posts
        USER_STATE.browse_count = USER_STATE.browse_count + 1
        
        if USER_STATE.browse_count >= 3 then
            USER_STATE.state = STATES.CREATE
        end
        
        local page = math.random(1, 5)
        return wrk.format("GET", "/api/v1/posts?page=" .. page .. "&page_size=20", {
            ["Authorization"] = "Bearer " .. (TOKEN or "")
        }, nil)
        
    elseif state == STATES.CREATE then
        -- User creates a post
        USER_STATE.state = STATES.READ
        
        local body = string.format([[
{
    "title": "User Journey Post %d",
    "content": "Created during load test",
    "status": "published"
}]], math.random(1, 10000))
        
        return wrk.format("POST", "/api/v1/posts", {
            ["Content-Type"] = "application/json",
            ["Authorization"] = "Bearer " .. (TOKEN or "")
        }, body)
        
    else
        -- User reads specific post, then starts over
        USER_STATE.state = STATES.BROWSE
        USER_STATE.browse_count = 0
        
        local post_id = math.random(1, 100)
        return wrk.format("GET", "/api/v1/posts/" .. post_id, {
            ["Authorization"] = "Bearer " .. (TOKEN or "")
        }, nil)
    end
end

-- Handle response
function response(status, headers, body)
    -- Extract token from login response
    if USER_STATE.state == STATES.BROWSE and TOKEN == nil then
        local token_match = body:match('"access_token":"([^"]+)"')
        if token_match then
            TOKEN = token_match
        end
    end
    
    if status >= 400 then
        print(string.format("Error: %d - %s", status, body:sub(1, 100)))
    end
end

-- Summary statistics
function done(summary, latency, requests)
    io.write("------------------------------\n")
    io.write("  User Journey Load Test Results\n")
    io.write("------------------------------\n")
    io.write(string.format("  Total requests: %d\n", summary.requests))
    io.write(string.format("  Total duration: %.2fs\n", summary.duration / 1000000))
    io.write(string.format("  Requests/sec:   %.2f\n", summary.requests / (summary.duration / 1000000)))
    io.write(string.format("  Avg latency:    %.2fms\n", latency.mean / 1000))
    io.write(string.format("  Max latency:    %.2fms\n", latency.max / 1000))
    io.write(string.format("  Total errors:   %d\n", summary.errors.connect + summary.errors.read + summary.errors.write + summary.errors.timeout))
    io.write("------------------------------\n")
end
