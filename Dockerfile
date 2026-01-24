FROM gocv/opencv:4.13.0

WORKDIR /app
COPY . .

RUN go build -o main .

CMD ["./main"]