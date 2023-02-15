FROM docker.rexcreation.net/opencv-yolo:latest AS build

COPY . /usr/local/go/src/app
WORKDIR /usr/local/go/src/app

RUN go build -o /app main.go

FROM docker.rexcreation.net/opencv-yolo:latest
COPY --from=build /app /app
CMD ["/app"]