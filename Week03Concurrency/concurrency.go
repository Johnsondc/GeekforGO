//1. 基于 errgroup 实现一个 http server 的启动和关闭 ，以及 linux signal 信号的注册和处理，要保证能够一个退出，全部注销退出。
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
)

//启动 HTTP server
func startHttpSvr(svr *http.Server) error {
	http.HandleFunc("/hello", aServer)
	fmt.Println("http server start")
	err := svr.ListenAndServe()
	return err
}

// 增加一个 HTTP hanlder
func aServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, world!\n")
}

func main() {
	ctx := context.Background()
	// 定义 withCancel -> cancel() 方法 去取消下游的 Context
	ctx, cancel := context.WithCancel(ctx)
	// 使用 errgroup 进行 goroutine 取消
	group, errCtx := errgroup.WithContext(ctx)
	//http server
	srv := &http.Server{Addr: ":9090"}

	group.Go(func() error {
		return startHttpSvr(srv)
	})

	group.Go(func() error {
		<-errCtx.Done() //阻塞。因为 cancel、timeout、deadline 都可能导致 Done 被 close
		fmt.Println("http server stop")
		return srv.Shutdown(errCtx)
	})

	chanel := make(chan os.Signal, 1) //这里要用 buffer 为1的 chan
	signal.Notify(chanel)

	group.Go(func() error {
		for {
			select {
			case <-errCtx.Done(): // 因为 cancel、timeout、deadline 都可能导致 Done 被 close
				return errCtx.Err()
			case <-chanel: // 因为 kill -9 或其他而终止
				cancel()
			}
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		fmt.Println("group error: ", err)
	}
	fmt.Println("all group done!")

}
//因为对go语言还很生疏,这里是参考网上其他同学写的,请老师多多包涵.
