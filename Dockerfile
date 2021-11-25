# Go build
FROM golang:alpine as build

WORKDIR /src/note-service/
COPY go.mod .
#COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o note-service ./cmd/service/

FROM alpine

COPY --from=build /src/note-service/note-service /note-service

ENTRYPOINT ["/note-service"]
