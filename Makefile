APPLICATION_NAME ?= simple_saver_service
CONTAINER_NAME ?= rest-server
PORT ?= 8080

run:
	docker build --tag ${APPLICATION_NAME} .
	docker run -d -p ${PORT}:${PORT} \
		-e AWS_ACCESS_KEY_ID=$$AWS_ACCESS_KEY_ID \
        -e AWS_SECRET_ACCESS_KEY=$$AWS_SECRET_ACCESS_KEY \
		-e AWS_REGION=$$AWS_REGION \
        --name ${CONTAINER_NAME} ${APPLICATION_NAME}

clean:
	docker stop ${CONTAINER_NAME}
	docker rm ${CONTAINER_NAME}
	docker image rm ${APPLICATION_NAME}

test:
	go test ./...