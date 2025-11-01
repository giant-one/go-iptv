FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN case "$TARGETARCH" in \
      amd64) cp ./license/license_amd64 ./license ;; \
      arm64) cp ./license/license_arm64 ./license ;; \
      arm)   cp ./license/license_armv7 ./license ;; \
      *) echo "未知架构: $TARGETARCH" && exit 1 ;; \
    esac

RUN  go build -o iptv main.go
RUN chmod +x /app/iptv

FROM alpine:latest

VOLUME /config
WORKDIR /app
EXPOSE 80 8080

ENV TZ=Asia/Shanghai
RUN apk add --no-cache openjdk8 bash curl tzdata sqlite;\
    cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
    
COPY ./client /client
COPY ./apktool/* /usr/bin/
COPY ./static /app/static
COPY ./database /app/database
COPY ./config.yml /app/config.yml
COPY ./README.md  /app/README.md
COPY ./logo /app/logo
COPY ./ChangeLog.md /app/ChangeLog.md
COPY ./Version /app/Version
COPY ./license /app/license

RUN chmod 777 -R /usr/bin/apktool* 

COPY --from=builder /app/iptv .

CMD ["./iptv"]