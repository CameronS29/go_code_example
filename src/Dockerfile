FROM golang

ENV VR_DB_HOST=$VR_DB_HOST
ENV VR_DB_PORT=$VR_DB_PORT
ENV VR_DB_USR=$VR_DB_USR
ENV VR_DB_PWD=$VR_DB_PWD
ENV VR_DB_NAME=$VR_DB_NAME
ENV VR_API_HOST=$VR_API_HOST
ENV VR_API_PORT=$VR_API_PORT

WORKDIR /usr/app

COPY . .

RUN go get github.com/lib/pq

EXPOSE 9090

CMD ["go", "run", "server.go"]