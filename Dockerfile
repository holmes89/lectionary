FROM golang:1.14-alpine AS base

FROM base as deps
WORKDIR "/lectionary"
ADD *.mod *.sum ./
RUN go mod download

FROM deps AS build-env
ADD cmd ./cmd
ADD internal ./internal
ADD resources ./resources
RUN go run cmd/load/main.go
ENV PORT 8080
EXPOSE 8080
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -X main.docker=true" -o server cmd/server/*.go
CMD ["./server"]

FROM scratch AS prod

WORKDIR /
ENV PORT 8080
EXPOSE 8080
COPY --from=build-env /lectionary/server /
COPY --from=build-env /lectionary/bible.db /
CMD ["/server"]