package main

import (
	"context"
	"demo/google"
	"demo/userip"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func handleSearch(w http.ResponseWriter, req *http.Request) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	timeout, err := time.ParseDuration(req.FormValue("timeout"))
	if err == nil {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	query := req.FormValue("q")
	if query == "" {
		http.Error(w, "no query", http.StatusBadRequest)
		return
	}

	userIP, err := userip.FromRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx = userip.NewContext(ctx, userIP)

	start := time.Now()
	results, err := google.Search(ctx, query)
	elapsed := time.Since(start)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := resultsTemplate.Execute(w, struct {
		Results          google.Results
		Timeout, Elapsed time.Duration
	}{
		Results: results,
		Timeout: timeout,
		Elapsed: elapsed,
	}); err != nil {
		log.Print(err)
		return
	}
}

var resultsTemplate = template.Must(template.New("results").Parse(`
<html>
<head/>
<body>
  <ol>
  {{range .Results}}
    <li>{{.Title}} - < a href=" ">{{.URL}}</ a></li>
  {{end}}
  </ol>
  <p>{{len .Results}} results in {{.Elapsed}}; timeout {{.Timeout}}</p >
</body>
</html>
`))

func producer(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, v := range nums {
			out <- v
		}
	}()
	return out
}

func square(inCh <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)

		for n := range inCh {
			out <- n * n
		}
	}()

	return out
}

func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)

	var wg sync.WaitGroup

	collect := func(in <-chan int) {
		defer wg.Done()
		for n := range in {
			out <- n
		}
	}

	wg.Add(len(cs))

	for _, c := range cs {
		go collect(c)
	}

	// 错误方式：直接等待是bug，死锁，因为merge写了out，main却没有读
	//wg.Wait()
	//close(out)

	go func() {
		fmt.Println("wait begin")
		time.Sleep(time.Second)
		close(out)
		wg.Wait()
		time.Sleep(5 * time.Second)
		fmt.Println("wait end")

	}()

	return out
}

func eat() chan string {
	out := make(chan string)
	go func() {
		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		out <- "mom call you eat"
		close(out)
	}()

	return out
}

type Bank struct {
	sync.RWMutex
	saving map[string]int
}

func (b *Bank) Deposit(name string, amount int) {
	b.Lock()
	defer b.Unlock()

	if _, ok := b.saving[name]; !ok {
		b.saving[name] = 0
	}

	b.saving[name] += amount
}

func (b *Bank) Withdraw(name string, amount int) int {
	b.Lock()
	defer b.Unlock()
	if _, ok := b.saving[name]; !ok {
		return 0
	}
	if b.saving[name] < amount {
		amount = b.saving[name]
	}
	b.saving[name] -= amount

	return amount
}

//查询余额
func (b *Bank) Query(name string) int {
	b.Lock()
	defer b.Unlock()

	if _, ok := b.saving[name]; !ok {
		return 0
	}

	return b.saving[name]
}

func NewBank() *Bank {
	b := &Bank{
		saving: make(map[string]int),
	}
	return b
}

func handle(wg *sync.WaitGroup, a int) chan int {
	out := make(chan int)
	go func() {
		time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
		out <- a
		wg.Done()
	}()

	return out
}

type IGreeting interface {
	sayHello()
}

func sayHello(i IGreeting) {
	i.sayHello()
}

type GO struct {
}

func (g GO) sayHello() {
	fmt.Println("hi i'm Go")
}

type PHP struct {
}

func (p PHP) sayHello() {
	fmt.Println("hi i'm PHP")
}

type Student struct {
	Name string
	Age  int
}

func (s Student) String() string {
	return fmt.Sprintf("[Name: %s], [Age: %d]", s.Name, s.Age)
}

func main() {
	var s = Student{
		Name: "qcrao",
		Age:  18,
	}

	fmt.Println(&s)
	//for ret:=range ch{
	//	fmt.Printf("%3d",ret)
	//}

	//fmt.Println(<-in)
	//http.HandleFunc("/search",handleSearch)
	//log.Fatal(http.ListenAndServe(":3000",nil))

	//var wg sync.WaitGroup
	//wg.Add(1)
	////NewTimer 创建一个 Timer，它会在最少过去时间段 d 后到期，向其自身的 C 字段发送当时的时间
	//timer1 := time.NewTimer(2 * time.Second)
	//NewTicker 返回一个新的 Ticker，该 Ticker 包含一个通道字段，并会每隔时间段 d 就向该通道发送当时的时间。它会调
	//整时间间隔或者丢弃 tick 信息以适应反应慢的接收者。如果d <= 0会触发panic。关闭该 Ticker 可
	//以释放相关资源。
	//ticker1 := time.NewTicker(2 * time.Second)

	//go func(t *time.Ticker) {
	//	defer wg.Done()
	//	for {
	//		<-t.C
	//		fmt.Println("get ticker1", time.Now().Format("2006-01-02 15:04:05"))
	//	}
	//}(ticker1)

	//go func(t *time.Timer) {
	//	defer wg.Done()
	//	//for {
	//		<-t.C
	//		fmt.Println("get timer", time.Now().Format("2006-01-02 15:04:05"))
	//		//Reset 使 t 重新开始计时，（本方法返回后再）等待时间段 d 过去后到期。如果调用时t
	//		//还在等待中会返回真；如果 t已经到期或者被停止了会返回假。
	//		//t.Reset(2 * time.Second)
	//	//}
	//}(timer1)
	//
	//wg.Wait()
	//time.Sleep(time.Second*10)

	//http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
	//	fmt.Println(request)
	//	writer.Write([]byte("hello goland"))
	//})
	//http.ListenAndServe(":3000",nil)
}
