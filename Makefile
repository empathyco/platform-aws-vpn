.PHONY: build clean gomodgen

build: gomodgen
	export GO111MODULE=on
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/api-authorizer 		github.com/empathyco/aws-vpn/cmd/lambda-api-authorizer
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/api-client 			github.com/empathyco/aws-vpn/cmd/lambda-api-client
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/api-server 			github.com/empathyco/aws-vpn/cmd/lambda-api-server
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/cert-stream 			github.com/empathyco/aws-vpn/cmd/lambda-cert-stream
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/revocation-notifier 	github.com/empathyco/aws-vpn/cmd/lambda-revocation-notifier
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/rotate-ca 				github.com/empathyco/aws-vpn/cmd/lambda-rotate-ca
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/ovpn-helper 			github.com/empathyco/aws-vpn/cmd/ovpn-helper
clean:
	rm -rf ./bin ./vendor Gopkg.lock

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh
