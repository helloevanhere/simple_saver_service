APPLICATION_NAME ?= docker-gs-ping
CONTAINER_NAME ?= rest-server
PORT ?= 8080

run:
	docker build --tag ${APPLICATION_NAME} .
	docker run -d -p ${PORT}:${PORT} --name ${CONTAINER_NAME} ${APPLICATION_NAME}

clean:
	docker stop ${CONTAINER_NAME}
	docker rm ${CONTAINER_NAME}
	docker image rm ${APPLICATION_NAME}