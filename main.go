package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)
type clientWithId struct {
	client *http.Client
	id int
}
type HTTPConnectionPool struct {
	mu sync.Mutex
	pool chan clientWithId
	connCount int
	maxSize int
}

func ConnectionFactory(connCount int) *clientWithId {
	return &clientWithId{
		client: &http.Client{},
		id: connCount,
	}
}

func NewPool(size int) *HTTPConnectionPool {
	connPool := &HTTPConnectionPool{
		maxSize: size,
		pool: make(chan clientWithId, size),
	}
	return connPool
}

func (pool *HTTPConnectionPool) Acquire() (clientWithId,error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	select {
	case conn := <- pool.pool:
		pool.connCount++
		return conn,nil;
	default:
		if pool.connCount < pool.maxSize {
			conn := ConnectionFactory(pool.connCount)
			fmt.Println("pool size ", pool.connCount)
			pool.connCount++
			return *conn,nil
		}
	}
	fmt.Println("waitiing ")
	return <- pool.pool,nil
}

func (pool *HTTPConnectionPool) Release(cid clientWithId) bool {	
	fmt.Println("releasing, pool conn count", pool.connCount)
	select {
	case pool.pool <- cid:
		pool.mu.Lock()
		defer pool.mu.Unlock()
		pool.connCount--
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
		conn.client.CloseIdleConnections()
	} 
}

func main() {
	pool := NewPool(5);
	// go server.Startserver()
	for i:=0 ; i< 10; i++ {
		go func(id int){
			conn,_ := pool.Acquire()
			// time.Sleep(3*time.Second)
			_, err := conn.client.Get("http://localhost:8080")
			if err == nil {
				fmt.Println("client acquired with id " , id)
			} else {
				fmt.Println("client not acquired, pool full", err)
			}
			fmt.Println("releasing....")
			pool.Release(conn)
		}(i)
	}
	// for i:=0;i<10;i++ {
	// 	_,err := pool.Acquire()
	// 	if err == nil {
	// 		fmt.Println("client acquired with id " , i)
	// 	} else {
	// 		fmt.Println("client not acquired, pool full")
	// 	}
	// 	// pool.Release(client)
	// }
	// fmt.Println("pool size ", len(pool.pool))
	time.Sleep(5*time.Second)
	pool.CloseAll()
}