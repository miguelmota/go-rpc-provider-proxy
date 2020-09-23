wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.body = '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}'
-- wrk.headers["X-Forwared-For"] = "127.0.0.1"

logfile = io.open("wrk.log", "w")

response = function(status, header, body)
  print(body)
  logfile:write("status:" .. status .. "\n" .. body .. "\n-------------------------------------------------\n");
end
