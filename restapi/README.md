# Go-iOS REST API

## getting started:
- Open up `restapi` folder in its own vscode window to start working on the api
- go install github.com/swaggo/swag/cmd/swag@latest
- swag init --parseDependency
- go run main.go

plug an ios device into your machine and test on localhost:8080

## structure
 - `api/routes.go`  contains all routes
 - `api/middleware.go` contains all middlewares
 - `api/*_endpoints.go` contains endpoints that mostly mirror go-ios docopt commands
 - `api/server.go` the server config


## to dos
APIs needed to solve automation problem, run WebDriverAgent with 0 hassle:
1. app install
2. dev image mount and check
3. run wda
4. wda shim/ tap and screenshot
5. signing api
6. wda binary download

