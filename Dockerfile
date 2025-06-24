FROM gocv/opencv:4.11.0

WORKDIR /app
COPY . .

RUN go build -o main .

CMD ["./main"]