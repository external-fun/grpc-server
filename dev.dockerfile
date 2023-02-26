FROM golang:1.18.1 as builder
ADD ./ /app/
WORKDIR /app
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /app/out/grpc-server .

FROM scratch
WORKDIR /app
COPY --from=builder /app/out/grpc-server /usr/bin/
EXPOSE 8080
EXPOSE 7070
CMD ["grpc-server"]