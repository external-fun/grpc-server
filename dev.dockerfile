FROM golang:alpine as builder
WORKDIR /app
ADD ./ /app/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/out/grpc-server .


FROM scratch
WORKDIR /app
COPY --from=builder /app/out/grpc-server /usr/bin/
EXPOSE 8080
ENTRYPOINT ["grpc-server"]