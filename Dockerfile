FROM golang:1.23 AS dependencies
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download


FROM dependencies AS build
COPY . .
WORKDIR /app
RUN make build

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /app/bin/app /app/
COPY /config/app.env /app/config/
COPY /scripts/wait-for-it.sh /app/
RUN chmod +x /app/app
EXPOSE 8080/tcp
CMD /app/app