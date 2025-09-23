
### Compile for lambda

```
go tool sqlc generate . 
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap -tags lambda.norpc main.go
zip golambda.zip bootstrap
```

### Tools added

SQLC 

```
go get -tool github.com/sqlc-dev/sqlc/cmd/sqlc
```

```
go tool sqlc generate .
```


### Linter

```
docker run -t --rm -v .:/app -w /app golangci/golangci-lint:v2.4.0 golangci-lint run
```

### Vulnerability scan

without login

```
docker run -t --rm -v .:/src semgrep/semgrep semgrep scan
```

TODO hot-swap function


The request response format uses the apigateway v2 format. 

```shell
curl -v 'https://abcdefg.lambda-url.us-east-1.on.aws/?message=HelloWorld' \
-H 'content-type: application/json' \
-d '{ "example": "test" }'
```


