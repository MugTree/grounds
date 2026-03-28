lint:
	golangci-lint run .

run-tests:
	cd www && go test -v && cd ..
  
format-html:
	templ fmt ./app 

dev:
	source ./scripts/load_env.sh && air  

production-build-app:
	templ generate && CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" -o ./bin/server .

vps-publish:
	./scripts/upload_site_to_vps.sh
