# 编译阶段
FROM golang:1.19-alpine AS builder

WORKDIR /app

# 1. 复制 go.mod (如果本地有 go.sum 也要复制，没有也没关系)
COPY go.mod ./

# --- 新增这一行 ---
# 自动下载并生成 go.sum，解决 checksum 报错
RUN go mod tidy
# -----------------

# 下载依赖
RUN go mod download

COPY main.go .
# 静态编译
RUN CGO_ENABLED=0 GOOS=linux go build -o honeypot .

FROM scratch

# Create the /tmp directory in the final image.
COPY --from=builder /tmp /tmp
COPY --from=builder /app/fakessh /fakessh
EXPOSE 2222
ENTRYPOINT ["/fakessh"]
