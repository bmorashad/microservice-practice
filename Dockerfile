FROM golang:1.18 as base

WORKDIR /app

COPY . .

RUN go mod tidy && go mod download

RUN go build -o /server

EXPOSE 8010

# CMD ["go", "run", "*.go"] 
CMD ["/server"] 
