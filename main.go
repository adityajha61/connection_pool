package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type HTTPConnectionPool struct {
	mu sync.Mutex
	pool chan *http.Client
	maxSize int
}

func ConnectionFactory() *http.Client {
	return &http.Client{};
}

func NewPool(size int) *HTTPConnectionPool {
	connPool := &HTTPConnectionPool{
		maxSize: size,
		pool: make(chan *http.Client, size),
	}
	for i :=0;i<size;i++{
		connPool.pool <- http.DefaultClient
	}
	return connPool
}

func (pool *HTTPConnectionPool) Acquire() (*http.Client,error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	select {
	case conn := <- pool.pool:
		return conn,nil;
	default:
		if len(pool.pool) < pool.maxSize {
			conn := ConnectionFactory()
			return conn,nil
		}
	}
	return <- pool.pool,nil
}

func (pool *HTTPConnectionPool) Release(conn *http.Client) bool {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	select {
	case pool.pool <- conn:
		return true;
	default:
		return false;
	}
}

func (pool *HTTPConnectionPool) CloseAll() {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	close(pool.pool)
	for conn:= range pool.pool {
		conn.CloseIdleConnections()
	} 
}

func main() {
	pool := NewPool(3);
	for i:=0 ; i< 10; i++ {
		go func(id int){
			client,err := pool.Acquire()

			if err == nil {
				fmt.Println("client acquired with id " , id)
			} else {
				fmt.Println("client not acquired, pool full")
			}
			pool.Release(client)
		}(i)
	}
	fmt.Println("pool size ", len(pool.pool))
	time.Sleep(1*time.Second)
	pool.CloseAll()
}