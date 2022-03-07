from golang

run mkdir /app

ADD . /app

WORKDIR /app

RUN go build .

EXPOSE 7777

CMD ["/app/location-service"]