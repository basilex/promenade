-- wrk Load Test: Posts CRUD Operations
-- Tests post creation with authenticated user
-- NOTE: Update TOKEN variable with valid access token before running

-- Configuration
local TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."  -- Replace with valid token
local counter = 0

-- Initialize
function init(args)
    if TOKEN:find("%.%.%.") then
        print("WARNING: Please update TOKEN variable with a valid access token")
    end
end

-- Generate request
function request()
    counter = counter + 1
    
    local operations = {"create", "list", "get"}
    local operation = operations[(counter % #operations) + 1]
    
    if operation == "create" then
        -- Create new post
        local body = string.format([[
{
    "title": "Load Test Post %d",
    "content": "This is a load test post created at %s",
    "status": "published"
}]], counter, os.date("%Y-%m-%d %H:%M:%S"))
        
        local path = "/api/v1/posts"
        local method = "POST"
        local headers = {
            ["Content-Type"] = "application/json",
            ["Authorization"] = "Bearer " .. TOKEN
        }
        
        return wrk.format(method, path, headers, body)
        
    elseif operation == "list" then
        -- List posts
        local path = "/api/v1/posts?page=1&page_size=20"
        local method = "GET"
        local headers = {
            ["Authorization"] = "Bearer " .. TOKEN
        }
        
        return wrk.format(method, path, headers, nil)
        
    else
        -- Get post by ID (random)
        local post_id = math.random(1, 100)
        local path = "/api/v1/posts/" .. post_id
        local method = "GET"
        local headers = {
            ["Authorization"] = "Bearer " .. TOKEN
        }
        
        return wrk.format(method, path, headers, nil)
    end
end

-- Handle response
function response(status, headers, body)
    if status ~= 200 and status ~= 201 and status ~= 404 then
        print(string.format("Error: %d - %s", status, body))
    end
end

-- Summary statistics
function done(summary, latency, requests)
    io.write("------------------------------\n")
    io.write("  Posts CRUD Load Test Results\n")
    io.write("------------------------------\n")
    io.write(string.format("  Total requests: %d\n", summary.requests))
    io.write(string.format("  Total duration: %.2fs\n", summary.duration / 1000000))
    io.write(string.format("  Requests/sec:   %.2f\n", summary.requests / (summary.duration / 1000000)))
    io.write(string.format("  Total errors:   %d\n", summary.errors.connect + summary.errors.read + summary.errors.write + summary.errors.timeout))
    io.write("------------------------------\n")
end
