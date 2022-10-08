# Go-iOS REST API

## getting started:
go install github.com/swaggo/swag/cmd/swag@latest
swag init
go run main.go

## structure
  api/routes.go  contains all routes
  api/middleware.go contains all middlewares
  api/*_endpoints.go contains endpoints that mostly mirror go-ios docopt commands
  api/server.go the server config
