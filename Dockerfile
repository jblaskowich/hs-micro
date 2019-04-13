FROM golang:1.11.5

LABEL maintainer="jblaskowichgmail.com"

WORKDIR /

RUN go get -v ./...

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM scratch  

COPY --from=0 app /

ENTRYPOINT ["/app"]