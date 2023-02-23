FROM golang:1.20-alpine

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

#Copy local source code
COPY . ./

# Build
RUN go build -o /simple_saver_service

# This is for documentation purposes only.
# To actually open the port, runtime parameters
# must be supplied to the docker command.
EXPOSE 8080

# Run
CMD [ "/simple_saver_service" ]
