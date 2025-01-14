FROM golang:1.20.5 as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

COPY main.go ./

RUN go build -o /router

FROM gcr.io/distroless/base-debian11

COPY --from=build /router /router

ENTRYPOINT [ "/router" ]
