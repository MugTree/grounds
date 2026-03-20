lint:
	golangci-lint run .

run-tests:
	cd www && go test -v && cd ..
  
format-html:
	templ fmt ./app 

templ:
	templ generate --watch --proxy="http://localhost:8080" --open-browser=false  -v

serve:
	air \
  --build.cmd "go build -o ./tmp/bin/app" --build.bin "./tmp/bin/app" --build.delay "100" \
  --build.include_ext "go,css,js" \
  --build.stop_on_error "false" \
  --misc.clean_on_exit true

sync_assets: 
	air \
  --build.cmd "templ generate --notify-proxy" \
  --build.bin "true" \
  --build.delay "100" \
  --build.exclude_dir "" \
  --build.include_dir "www,public" \
  --build.include_ext "js,css"

start-dev: 
	make  -j3 templ serve sync_assets
#	make -j 3  templ serve sync_assets

production-build-app:
	templ generate && CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" -o ./bin/server .

vps-publish:
	./scripts/upload_site_to_vps.sh
