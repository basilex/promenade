-- wrk Load Test: Notifications Module
-- Tests notification creation and preference management
-- NOTE: Update TOKEN variable with valid access token before running

-- Configuration
local TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."  -- Replace with valid token
local counter = 0
local notification_types = {"system", "security", "product", "social"}
local notification_channels = {"email", "sms", "push", "in_app"}

-- Initialize
function init(args)
    if TOKEN:find("%.%.%.") then
        print("WARNING: Please update TOKEN variable with a valid access token")
    end
    math.randomseed(os.time())
end

-- Setup thread
function setup(thread)
    thread:set("id", counter)
    counter = counter + 1
end

-- Generate request
function request()
    counter = counter + 1
    
    local operations = {"send", "list", "preferences", "unread_count"}
    local operation = operations[(counter % #operations) + 1]
    
    if operation == "send" then
        -- Send notification
        local notif_type = notification_types[math.random(1, #notification_types)]
        local notif_channel = notification_channels[math.random(1, #notification_channels)]
        
        local body = string.format([[
{
    "type": "%s",
    "channel": "%s",
    "subject": "Load Test Notification %d",
    "content": "This is a load test notification created at %s",
    "data": {
        "test": true,
        "counter": %d
    }
}]], notif_type, notif_channel, counter, os.date("%Y-%m-%d %H:%M:%S"), counter)
        
        local path = "/api/v1/notifications"
        local method = "POST"
        local headers = {
            ["Content-Type"] = "application/json",
            ["Authorization"] = "Bearer " .. TOKEN
        }
        
        return wrk.format(method, path, headers, body)
        
    elseif operation == "list" then
        -- List notifications
        local page = math.random(1, 5)
        local path = string.format("/api/v1/notifications?page=%d&page_size=20", page)
        local method = "GET"
        local headers = {
            ["Authorization"] = "Bearer " .. TOKEN
        }
        
        return wrk.format(method, path, headers, nil)
        
    elseif operation == "preferences" then
        -- Get preferences
        local path = "/api/v1/notifications/preferences"
        local method = "GET"
        local headers = {
            ["Authorization"] = "Bearer " .. TOKEN
        }
        
        return wrk.format(method, path, headers, nil)
        
    else
        -- Get unread count
        local path = "/api/v1/notifications/unread-count"
        local method = "GET"
        local headers = {
            ["Authorization"] = "Bearer " .. TOKEN
        }
        
        return wrk.format(method, path, headers, nil)
    end
end

-- Response handler
function response(status, headers, body)
    if status ~= 200 and status ~= 201 then
        print("ERROR: Status " .. status)
        print("Body: " .. body)
    end
end

-- Statistics
function done(summary, latency, requests)
    io.write("------------------------------\n")
    io.write("Notifications Stress Test Results\n")
    io.write("------------------------------\n")
    io.write(string.format("Total Requests:    %d\n", summary.requests))
    io.write(string.format("Total Duration:    %.2fs\n", summary.duration / 1000000))
    io.write(string.format("Requests/sec:      %.2f\n", summary.requests / (summary.duration / 1000000)))
    io.write(string.format("Total Errors:      %d\n", summary.errors.connect + summary.errors.read + summary.errors.write + summary.errors.status + summary.errors.timeout))
    io.write(string.format("Error Rate:        %.2f%%\n", (summary.errors.connect + summary.errors.read + summary.errors.write + summary.errors.status + summary.errors.timeout) / summary.requests * 100))
    io.write("\n")
    io.write("Latency Distribution:\n")
    io.write(string.format("  50%%:            %.2fms\n", latency:percentile(50)))
    io.write(string.format("  75%%:            %.2fms\n", latency:percentile(75)))
    io.write(string.format("  90%%:            %.2fms\n", latency:percentile(90)))
    io.write(string.format("  99%%:            %.2fms\n", latency:percentile(99)))
    io.write(string.format("  Max:             %.2fms\n", latency.max))
    io.write("------------------------------\n")
end
