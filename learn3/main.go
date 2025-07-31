package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Task struct {
	Conn net.Conn
	Data string
}

const NumWorkers = 10
const TaskQueueSize = 100

var TaskQueue = make(chan Task, TaskQueueSize)

func handleConnectionForPool(conn net.Conn) {
	// 客人离开时，确保连接被关闭
	defer conn.Close()

	log.Printf("一位新客人 %s 到来，服务员准备为其点菜。", conn.RemoteAddr())

	// 使用 bufio.Reader 可以方便地读取一行数据
	reader := bufio.NewReader(conn)
	for {
		// 等待客人点菜（客户端发送一行数据）
		// ReadString('\n') 会阻塞直到读到换行符
		data, err := reader.ReadString('\n')
		if err != nil {
			// 如果读取出错（通常是连接断开），服务员就结束对这位客人的服务
			log.Printf("客人 %s 已离开。", conn.RemoteAddr())
			return
		}

		// 把客人的要求写成一张标准订单
		task := Task{
			Conn: conn,
			Data: data,
		}

		// 服务员尝试把订单贴到订单墙上
		TaskQueue <- task
		log.Printf("服务员已将来自 %s 的订单成功贴到墙上。", conn.RemoteAddr())
	}
}

// 每个厨师都是一个独立的goroutine，执行这个函数
// id 是厨师的工号
// wg 是一个同步工具，我们稍后解释
func worker(id int, wg *sync.WaitGroup) {
	// defer 语句确保在函数退出时，一定会执行 wg.Done()
	defer wg.Done()

	log.Printf("厨师 %d 已就位，准备开工！", id)

	// for...range 用于channel时，会一直阻塞等待，直到channel中有新的任务到来
	// 或者channel被关闭。这是一个非常常见的模式。
	for task := range TaskQueue {
		log.Printf("厨师 %d 接到来自 %s 的订单，开始炒菜: %s", id, task.Conn.RemoteAddr(), task.Data)

		// 模拟一个非常耗时的炒菜过程
		time.Sleep(3 * time.Second)

		// 菜做好了
		result := fmt.Sprintf("'炒好的菜品：%s' (由厨师 %d 制作)\n", strings.TrimSpace(task.Data), id)

		// 把菜送回给客人
		task.Conn.Write([]byte(result))
	}

	log.Printf("厨师 %d 的订单墙空了，并且墙被拆了(channel关闭)，下班！", id)
}
func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("餐厅开张失败: %v", err)
	}
	defer listener.Close()
	log.Println("餐厅开张，正在监听 :8080 端口...")

	var wg sync.WaitGroup // 准备好下班打卡器

	// 雇佣厨师，让他们去后厨准备
	for i := 1; i <= NumWorkers; i++ {
		wg.Add(1) // 每雇佣一个，打卡器上+1
		go worker(i, &wg)
	}

	// 开始接待客人
	for {
		conn, err := listener.Accept() // 等待新客人进门
		if err != nil {
			log.Printf("接待客人失败: %v", err)
			continue
		}
		// 每来一位客人，就派一位新的服务员（goroutine）去专门为他服务
		go handleConnectionForPool(conn)
	}

	// --- 优雅关闭的逻辑 (在实际应用中) ---
	// (例如，捕获 Ctrl+C 信号后执行)
	// log.Println("餐厅准备关门...")
	// close(taskQueue) // 拆掉订单墙，厨师们做完手上的单就不能再接了
	// wg.Wait() // 经理等待所有厨师都打卡下班
	// log.Println("所有厨师已下班，餐厅正式关闭。")
}
