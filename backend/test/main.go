package test

import (
	"blog-for-go/redis"
	"blog-for-go/util"
	"context"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

var rdbCtx = context.Background()

func main() {
	app := iris.New()

	tmpl := iris.HTML("./templates", ".html")
	tmpl.Reload(true) // reload templates on each request (development mode)
	app.RegisterView(tmpl)

	app.OnErrorCode(iris.StatusNotFound, func(ctx iris.Context) {
		ctx.View("404.html")
	})

	app.Get("/", func(ctx iris.Context) {
		ctx.ViewData("Title", "Index Page Of Go(Iris)")
		ctx.ViewData("Msg", "这是一条测试消息！")
		ctx.View("index.html")
	})
	app.Get("/cpu", func(ctx iris.Context) {
		const path = "/sys/class/thermal/thermal_zone0/temp"
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read cpu temperature form %q:%v", path, err)
		}

		cpuTempStr := strings.TrimSpace(string(data))
		cpuTempInt, err := strconv.Atoi(cpuTempStr)

		if err != nil {
			log.Fatalf("%q does not contains an integer: %v", path, err)
		}

		cpuTemp := float64(cpuTempInt) / 1000.0
		ctx.ViewData("CpuTemp", cpuTemp)

		ctx.View("cpu.html")
	})

	app.Get("/greeting/{userName}", func(ctx iris.Context) {
		ctx.Writef("hello, %s", ctx.Params().Get("userName"))
	})

	// 注册函数
	app.Macros().Get("int").RegisterFunc("min", func(minVal int) func(string) bool {
		return func(paramVal string) bool {
			n, err := strconv.Atoi(paramVal)

			if err != nil {
				return false
			}

			return n >= minVal
		}
	})
	app.Get("/testMin/{id:int min(1) else 504}", func(ctx iris.Context) {
		id, _ := ctx.Params().GetInt("id")
		ctx.Writef("Hello id: %d", id)
	})

	app.Get("/game/{name:alphabetical}/level/{level:int}", func(ctx iris.Context) {
		ctx.Writef("name: %s | level: %s", ctx.Params().Get("name"), ctx.Params().Get("level"))
	})

	//app.Get("/single_file/{myfile:file}", func(ctx iris.Context) {
	//	ctx.Writef("file type validates if the parameter value has a form of a file name, got: %s", ctx.Params().Get("myfile"))
	//})

	app.Get("/myfiles/{directory:path}", func(ctx iris.Context) {
		ctx.Writef("path type accepts any number of path segments, path after /myfiles/ is: %s", ctx.Params().Get("directory"))
	})

	app.Get("/redis/{mode}/{key}/{value}", func(ctx iris.Context) {
		mode := ctx.Params().Get("mode")
		key := ctx.Params().Get("key")
		value := ctx.Params().Get("value")

		rdb := redis.GetRedisClient(mode)
		set := rdb.Set(rdbCtx, key, value, 200)

		ctx.ViewData("Key", key)
		ctx.ViewData("Err", set.Err())
		ctx.ViewData("Val", set.Val())
		ctx.ViewData("Rs", util.PrettyJSON(set))

		ctx.View("redis.html")
	})

	app.Get("/redis/{mode}/{key}", func(ctx iris.Context) {
		mode := ctx.Params().Get("mode")
		key := ctx.Params().Get("key")

		log.Println("mode:" + mode)
		log.Println("key:" + key)

		rdb := redis.GetRedisClient(mode)

		value := rdb.Get(rdbCtx, key)
		log.Println("value:", value.Val())

		ctx.ViewData("Key", key)
		ctx.ViewData("Err", value.Err())
		ctx.ViewData("Rs", util.PrettyJSON(value))

		ctx.View("redis.html")
	})

	// controllers 演示
	mvc.New(app.Party("/hello")).Handle(new(ExpController))

	app.Run(iris.Addr(":8080"))
}

type ExpController struct {
}

type ResponseData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json: "data"`
}

func GetSuccessData(data interface{}) *ResponseData {
	return &ResponseData{200, "hello", data}
}

func (c *ExpController) Get() mvc.Result {
	return mvc.View{
		Name: "index.html",
		Data: map[string]interface{}{
			"Title": "hello",
			"Msg":   "message",
		},
	}
}

func (c *ExpController) GetJson() interface{} {
	return GetSuccessData("这是一条数据")
}
