FROM golang:1.19
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN mv .env.template .env
RUN GOOS=linux go build -v
CMD ["./syncd"]
