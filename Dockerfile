FROM golang:latest

WORKDIR app 

# COPY go.mod go.sum ./
COPY . ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server.exe server/main.go

EXPOSE 7000:7000

CMD ["/bin/server.exe" ]
