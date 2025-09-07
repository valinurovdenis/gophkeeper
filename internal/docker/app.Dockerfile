FROM golang:alpine

WORKDIR /app
COPY . .
RUN go build -o gophkeeper-server ./server
EXPOSE 8080
CMD ["/app/gophkeeper-server"]