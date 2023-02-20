FROM golang:1.19-alpine AS build
WORKDIR /app
COPY . .
RUN apk add --no-cache make gcc musl-dev linux-headers
RUN make build-cli && make build-miner

FROM scratch
COPY --from=build /app/bin/ /app/bin/
RUN export WALLET_ADDR=$(/app/bin/cli wallet create | grep "address" | awk '{print $5}')
ENV WALLET_ADDR=$WALLET_ADDR
EXPOSE 1234
ENTRYPOINT ["/app/bin/miner"]
