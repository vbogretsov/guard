FROM golang:1.16-alpine as build
ARG VERSION
WORKDIR /out
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s -X 'main.Version=${VERSION}'" ./cmd/guard
RUN apk add upx
RUN upx guard
RUN adduser -D -H -S guard

FROM scratch
COPY --from=build /out/guard /bin/guard
COPY --from=build /etc/passwd /etc/passwd
USER guard
ENTRYPOINT [ "/bin/guard" ]
