# from alpine image
FROM golang:1.16-alpine

# set root user
USER root

# app work directory
WORKDIR /app

# copy mod and sum
COPY go.mod go.sum ./

# now we download dependencies
RUN go mod download

# then we copy project files into the app directory
COPY . ./

# building the application
RUN go build -o /main cmd/main.go

# running application
ENTRYPOINT ./main
