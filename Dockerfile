FROM golang:1.12.5 AS builder
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make build

FROM gcr.io/distroless/base
COPY --from=builder /src/build/cloudsql-postgres-operator /cloudsql-postgres-operator
CMD ["/cloudsql-postgres-operator", "-h"]
