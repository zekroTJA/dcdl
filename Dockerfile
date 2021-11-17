FROM golang:1.17-alpine AS build
WORKDIR /build
COPY . .
RUN go mod download
RUN go build -o bin/dcdl cmd/dcdl/main.go

FROM alpine:latest
COPY --from=build /build/bin/dcdl /bin/dcdl
ENV DCDL_WEBSERVER_BINDADDRESS="0.0.0.0:80"
ENV DCDL_STORAGE_LOCATION="/var/data"
RUN mkdir -p ${DCDL_STORAGE_LOCATION}
EXPOSE 80
ENTRYPOINT ["/bin/dcdl"]