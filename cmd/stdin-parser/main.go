package main

import (
	//"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Unknwon/goconfig"
	"github.com/chengongliang/mysqllog"
	"github.com/hpcloud/tail"
	"io/ioutil"
	"net/http"
	"strconv"
)

type CONF struct {
	Token   string
	LogFile string
	QTime   string
}

func SendDingTalk(message, token string) {
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
	p := &mysqllog.Parser{}
	cfg, err := goconfig.LoadConfigFile("conf.ini")
	if err != nil {
		panic("未找到 conf.ini 文件.")
	}
	var c CONF
	c.Token, err = cfg.GetValue("dingTalk", "token")
	if err != nil {
		fmt.Println("dintTalk -> token 未配置")
	}
	c.LogFile, err = cfg.GetValue("log", "path")
	if err != nil {
		fmt.Println("log -> path 未配置")
	}
	c.QTime, err = cfg.GetValue("log", "query_time")
	if err != nil {
		fmt.Println("log -> query_time 未配置")
	}
	qTime, _ := strconv.ParseFloat(c.QTime, 32/64)
	t, err := tail.TailFile(c.LogFile, tail.Config{Follow: true})
	if err != nil {
		fmt.Println(err)
	}
	for line := range t.Lines {
		event := p.ConsumeLine(line.Text)
		if event != nil && len(event) != 0 {
			if event["Query_time"] == nil {
				continue
			}
			if event["Query_time"].(float64) > qTime {
				msg := fmt.Sprintf("# <font face=\"微软雅黑\">慢SQL通知</font>\n \n <br/> \n **数据库:** %v\n\n<br/>**IP:** %v\n\n<br/>**SQL 时间:** %v\n\n<br/>**执行时间:** %v\n\n<br/>**执行内容:** %v", event["Schema"], event["IP"], event["Timestamp"], event["Query_time"], event["Statement"])
				SendDingTalk(msg, c.Token)
			}
		}
	}
}
