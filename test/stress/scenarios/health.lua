-- wrk Load Test: Health Endpoints
-- Tests all module health endpoints under load

-- Configuration
local counter = 0
local modules = {
    "/api/v1/health",
    "/api/v1/posts/health",
    "/api/v1/profiles/health",
    "/api/v1/analytics/health",
    "/api/v1/notifications/health",
    "/api/v1/billing/health"
}

-- Initialize
function init(args)
    math.randomseed(os.time())
end

-- Setup thread
function setup(thread)
    thread:set("id", counter)
    counter = counter + 1
end

-- Generate request
function request()
    -- Randomly select a health endpoint
    local idx = math.random(1, #modules)
    local path = modules[idx]
    
    return wrk.format("GET", path, {
        ["Content-Type"] = "application/json"
    })
end

-- Response handler
function response(status, headers, body)
    if status ~= 200 then
        print("ERROR: Status " .. status .. " for health endpoint")
        print("Body: " .. body)
    end
end

-- Statistics
function done(summary, latency, requests)
    io.write("------------------------------\n")
    io.write("Health Endpoints Stress Test Results\n")
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
