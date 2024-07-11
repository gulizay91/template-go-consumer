.PHONY: run_docker stop_docker docker_latest_image

run_docker:
	docker rm -f template-go-consumer || true && \
        docker rmi -f $(docker images -q template-go-consumer) || true && \
    		docker build --build-arg SERVICE_PORT=7011 -t template-go-consumer ./app && \
    			docker run --rm --name template-go-consumer -p 7011:7011 -e SERVICE__ENVIRONMENT=dev -e SERVICE__PORT=7011 -d template-go-consumer

stop_docker:
	docker stop template-go-consumer

docker_latest_image:
	docker build -t template-go-consumer ./app && \
		docker tag template-go-consumer guliz91/template-go-consumer:latest && \
			docker push guliz91/template-go-consumer:latest