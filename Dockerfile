# 编译阶段
FROM golang:1.19-alpine AS builder

WORKDIR /app

# 1. 先复制依赖定义文件
COPY go.mod ./

# 2. 复制源代码
COPY main.go .

# 3. 缺包并自动下载
RUN go mod tidy

# 4. 再次确保下载完成
RUN go mod download

# 5. 静态编译
RUN CGO_ENABLED=0 GOOS=linux go build -o honeypot .

# 运行阶段
FROM alpine:latest

WORKDIR /app
# 创建日志目录
RUN mkdir /logs

# 从编译阶段复制二进制文件
COPY --from=builder /app/honeypot .

# 运行
CMD ["./honeypot"]
