package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var client = &http.Client{}
var referer string

func getData(url string) (*[]byte, string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Mobile Safari/537.36")
	req.Header.Set("referer", referer)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("404 请求失败！")
		return nil, "404"
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return &body, resp.Status[:3]
}

// 下载计数
var count int

func saveImage(urlList *[]string, dir string) {
	for i, v := range *urlList {
		path := dir + fmt.Sprintf("%03d.jpg ", i + 1)
		count++
		fmt.Printf("%d %s", count, path)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("已下载\n")
			continue
		}
		resp, _ := getData(v)
		// 空数据就不写入
		if *resp == nil {
			continue
		}
		f, err := os.Create(path)
		if err != nil {
			fmt.Println("图片保存失败！")
			continue
		}
		defer f.Close()
		f.Write(*resp)
		fmt.Println("下载成功")
	}
}

// 检测版本
func init() {
	const version = "v1.0.2"
	url := "https://docs.qq.com/dop-api/opendoc?id=DT3F6UmhxS3VaQXZ1&normal=1"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("版本检测错误！")
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	re := regexp.MustCompile("loli{(.+?),(.+?),(.+?)}loli")
	res := re.FindAllSubmatch(body, -1)
	ver := string(res[0][1])
	updateContent := string(res[0][2])
	link := string(res[0][3])
	if ver != version {
		fmt.Printf("当前版本：%s\n最新版本：%s\n更新内容：%s\n%s\n", version, ver, updateContent, link)
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

func main() {
	var url string
	// 匹配分类 标题 图集链接数据
	re := regexp.MustCompile(`rel="category">#(.+?)#</a></div>[\s\S]+?_blank">(.+?)</a></h2>[\s\S]+?<div class="uk-inline">(.+?)\n`)
	// 匹配图集链接
	reUrl := regexp.MustCompile(`(https.+?\.jpg)`)
	page := 1
	// 路径分隔符
	pathSeparator := filepath.FromSlash("/")
	path := filepath.Dir(os.Args[0]) + pathSeparator + "image" + pathSeparator
	for {
		if page == 1 {
			url = "https://mmzztt.com/beauty/"
			referer = url
		} else {
			referer = url
			url = "https://mmzztt.com/beauty/page/" + fmt.Sprintf("%d", page)
		}
		resp, status := getData(url)
		if status != "200" {
			fmt.Println("下载完毕！")
			return
		}
		page++
		res := re.FindAllSubmatch(*resp, -1)
		for _, v := range res {
			dir := path + string(v[1]) + pathSeparator + strings.Replace(string(v[2]), ".", "", -1) + pathSeparator
			// 创建文件夹
			if _, err := os.Stat(dir); err != nil {
				os.MkdirAll(dir, 0777)
			}
			// 获取图集链接
			resUrl := reUrl.FindAllSubmatch(v[3], -1)
			urlList := make([]string, len(resUrl))
			for i, k := range resUrl {
				urlList[i] = strings.Replace(string(k[1]), "thumb300", "mw2000", -1)
			}
			// 保存图片
			saveImage(&urlList, dir)
			// 延时3秒
			time.Sleep(3 * time.Second)
		}
	}
}