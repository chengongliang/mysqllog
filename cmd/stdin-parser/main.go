package main

import (
	//"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/chengongliang/mysqllog"
	"github.com/hpcloud/tail"

	"github.com/Unknwon/goconfig"
)

type CONF struct {
	Token   string
	QTime   string
	WhiteIP string
	Log     string
}

func sendDingTalk(message, token string) {
	type cnt struct {
		Text  string `json:"text"`
		Title string `json:"title"`
	}

	type msg struct {
		MsgType  string `json:"msgtype"`
		Markdown cnt    `json:"markdown"`
	}

	url := "https://oapi.dingtalk.com/robot/send?access_token=" + token
	text := msg{}
	text.MsgType = "markdown"
	text.Markdown.Text = message
	text.Markdown.Title = "SQL报警"
	t, _ := json.Marshal(text)
	r, err := http.Post(url, "application/json;charset=utf-8", bytes.NewReader([]byte(t)))
	if err != nil {
		fmt.Println(err)
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(body))
}

func main() {
	//cfg, err := goconfig.LoadConfigFile("/Users/chengongliang/go/src/github.com/chengongliang/mysqllog/cmd/stdin-parser/conf.ini")
	cfg, err := goconfig.LoadConfigFile("conf.ini")
	if err != nil {
		panic("未找到 conf.ini 文件.")
	}
	var c CONF
	c.Token, err = cfg.GetValue("dingTalk", "token")
	if err != nil {
		fmt.Println("dintTalk -> token 未配置")
	}
	c.Log, _ = cfg.GetValue("logs", "path")
	if err != nil {
		fmt.Println("logs -> path 未配置")
	}
	c.QTime, err = cfg.GetValue("base", "query_time")
	if err != nil {
		fmt.Println("base -> query_time 未配置")
	}
	c.WhiteIP, err = cfg.GetValue("base", "white_ip")
	if err != nil {
		fmt.Println("base -> white_ip 未配置")
	}
	type Target struct {
		Addr    string `json:"addr"`
		LogFile string `json:"log_file"`
	}
	var target []Target
	er := json.Unmarshal([]byte(c.Log), &target)
	if er != nil {
		fmt.Println(err)
	}
	var wg sync.WaitGroup
	for _, v := range target {
		wg.Add(1)
		go func(v Target) {
			fmt.Println("开始监控: ", v.LogFile)
			p := &mysqllog.Parser{}
			qTime, _ := strconv.ParseFloat(c.QTime, 32/64)
			t, err := tail.TailFile(v.LogFile, tail.Config{Follow: true})
			if err != nil {
				fmt.Println(err)
			}
			for line := range t.Lines {
				event := p.ConsumeLine(line.Text)
				if event != nil && len(event) != 0 {
					if event["Query_time"] == nil || strings.Contains(c.WhiteIP, event["IP"].(string)) {
						continue
					}
					if event["Query_time"].(float64) > qTime {
						msg := fmt.Sprintf("# <font face=\"微软雅黑\">慢SQL通知</font>\n\n<br/>\n**地址:** %v\n\n<br/>**DB:** %v\n\n<br/>**来源IP:** %v\n\n<br/>**SQL 时间:** %v\n\n<br/>**执行时间:** %v\n\n<br/>**执行内容:** ```%v```",
							v.Addr, event["Schema"], event["IP"], event["Timestamp"], event["Query_time"], event["Statement"])
						sendDingTalk(msg, c.Token)
						//fmt.Println(msg)
					}
				}
			}
			fmt.Println("end")
			defer wg.Done()
		}(v)
	}
	wg.Wait()
}
