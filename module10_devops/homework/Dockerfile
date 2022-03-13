FROM golang:1.17-alpine
ENV GOPROXY="https://goproxy.cn"
COPY ./httpserver  /go/src/httpserver
WORKDIR /go/src/httpserver/
RUN go build -o /bin/httpserver
ENTRYPOINT ["/bin/httpserver"]
EXPOSE 80
