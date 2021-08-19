FROM golang:1.16.2-alpine3.13
RUN apk add ca-certificates && update-ca-certificates
WORKDIR /build
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o notes

FROM alpine:3.14.1
COPY --from=0 /build/dist /dist
COPY --from=0 /build/notes /notes
EXPOSE 8000
CMD ["/notes"]