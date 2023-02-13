FROM docker.rexcreation.net/opencv-yolo:latest AS build

ARG BITBUCKET_READ_TOKEN
ENV GOPRIVATE="bitbucket.org/arwineap"

COPY . /usr/local/go/src/app
WORKDIR /src

RUN go build -o /app main.go

FROM docker.rexcreation.net/opencv-yolo:latest
COPY --from=build /app /app
CMD ["./app"]