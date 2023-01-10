FROM gocv/opencv:4.7.0

WORKDIR /app
COPY . .

RUN go build -o main .

CMD ["./main"]