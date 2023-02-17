FROM golang as builder

RUN mkdir /src
RUN go env -w GOPROXY="https://goproxy.cn,direct"
COPY . /src

RUN cd /src && go build -o redis-transfer

FROM busybox

COPY --from=builder /src/redis-transfer /redis-transfer

ENTRYPOINT /redis-transfer
