package search

import (
	"log"
	"sync"
)

// A map of registered matchers for searching. 用于搜索的注册匹配器的map
var matchers = make(map[string]Matcher)

// Run performs the search logic. Run执行搜索逻辑
func Run(searchTerm string) {
	// Retrieve the list of feeds to search through. 获取需要搜索的数据源列表.
	feeds, err := RetrieveFeeds()
	if err != nil {
		log.Fatal(err)
	}

	// Create an unbuffered channel to receive match results to display. 创建一个无缓冲的通道,接受匹配后的结果
	results := make(chan *Result)

	// Setup a wait group so we can process all the feeds. 构造一个waitGroup,以便处理所有的数据
	var waitGroup sync.WaitGroup

	// Set the number of goroutines we need to wait for while 设置需要等待处理
	// they process the individual feeds.每个数据源的goroutine的数量
	waitGroup.Add(len(feeds))

	// Launch a goroutine for each feed to find the results.为每一个数据源启动一个goroutine来查找数据
	for _, feed := range feeds {
		// Retrieve a matcher for the search. 获取设置一个匹配器用查找数据
		matcher, exists := matchers[feed.Type]
		if !exists {
			matcher = matchers["default"]
		}

		// Launch the goroutine to perform the search. 启动一个goroutine来执行搜索
		go func(matcher Matcher, feed *Feed) {
			Match(matcher, feed, searchTerm, results)
			waitGroup.Done()
		}(matcher, feed)
	}

	// Launch a goroutine to monitor when all the work is done. 启动一个goroutine来监控是否所有的工作都做完了
	go func() {
		// Wait for everything to be processed. 等待所有的任务都完成
		waitGroup.Wait()

		// Close the channel to signal to the Display 这里我们使用关闭通道额方式通知display函数
		// function that we can exit the program. 这里就直接退出程序
		close(results)
	}()

	// Start displaying results as they are available and 启动函数,显示返回结果,并且
	// return after the final result is displayed.在最后一个结果显示完成后的返回
	Display(results)
}

// Register is called to register a matcher for use by the program.
func Register(feedType string, matcher Matcher) {
	if _, exists := matchers[feedType]; exists {
		log.Fatalln(feedType, "Matcher already registered")
	}

	log.Println("Register", feedType, "matcher")
	matchers[feedType] = matcher
}