# Simple (Proxy) Mock

## How to use (example)
1. Start proxy server in on e shell
go run spm.go 8080 "https://google.com"
2. In another shell make request to the proxy
curl http://localhost:8080/
3. Set mock for specific endpoint
curl -X POST -d "{\"endpoint\": \"/ping\", \"statusCode\": 200, \"body\": \"This is mocked pong\"}" http://localhost:8080/mockSettings/set
4. Send request to the endpoint
curl http://localhost:8080/ping
5. Clear mock request
curl -X POST -d "{\"endpoint\": \"/ping\"}" http://localhost:8080/mockSettings/clear
6. Try to make request to removed mock
curl http://localhost:8080/ping
