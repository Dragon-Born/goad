FROM golang:1.20.7-alpine3.18 as dev-env

RUN apk add --no-cache build-base opencv opencv-dev

COPY . /go/src/
WORKDIR /go/src/

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -gcflags "all=-N -l" -o /server ./main.go

FROM alpine:3.18
RUN apk add --no-cache opencv tzdata
RUN mkdir /data
WORKDIR /data

COPY --from=dev-env /server ./
COPY --from=dev-env /go/src/config.yaml ./
ENV TZ=Asia/Tehran
RUN cp /usr/share/zoneinfo/Asia/Tehran /etc/localtime
CMD ["./server"]