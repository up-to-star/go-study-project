.PHONY: docker
docker:
	@rm webook || true
	@GOOS=linux GOARCH=amd64 go build -tags=k8s -o webook .
	@docker rmi -f flycash/webook:v0.0.1
	@docker build -t flycash/webook:v0.0.1 .