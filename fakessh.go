package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// 配置日志轮转
	// 限制日志文件无限膨胀
	log.SetOutput(&lumberjack.Logger{
		Filename:   "/logs/access.log",
		MaxSize:    10,   // 每个日志文件最大 10MB
		MaxBackups: 3,    // 最多保留 3 个旧文件
		MaxAge:     28,   // 最多保留 28 天
		Compress:   true, // 旧日志是否压缩(gz)
	})

	// 设置日志格式，移除时间戳（fail2ban 处理更简单），或者保留并调整 regex
	// 只记录 IP，Fail2ban 根据系统时间处理
	log.SetFlags(0) 

	port := "2222"
	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("High-Performance Honeypot listening on port %s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %s\n", err)
			continue
		}
		// 使用协程处理，轻松应对每秒成千上万次扫描
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	ip, _, _ := net.SplitHostPort(remoteAddr)

	// 2. 写入日志文件 (Fail2ban 监控)
	// 格式: [HONEYPOT] <IP>
	log.Printf("[HONEYPOT] %s\n", ip)

	// 3. 发送假 Banner 并立即断开
	time.Sleep(10 * time.Millisecond)
	conn.Write([]byte("SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5\r\n"))
}
