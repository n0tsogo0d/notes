FROM golang:1.16.2-alpine3.13
WORKDIR /build
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o notes

FROM scratch
COPY --from=0 /build/web /web
COPY --from=0 /build/notes /notes
EXPOSE 8000
CMD ["/notes"]