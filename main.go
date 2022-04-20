package main

import (
	"MyWebhook/lib"
	"bytes"
	"os/exec"

	"github.com/gin-gonic/gin"
)

const (
	//元宇宙官网
	WEBSIET = 41
	//元宇宙api
	WEBSIETAPI = 42
	//元宇宙后台管理web
	ADMINWEB = 44
)

type Msg struct {
	Event_name string `form:"event_name" json:"event_name" binding:"required"`
	Project_id int    `form:"project_id" json:"project_id" binding:"required"`
	Ref        string `form:"ref" json:"ref" binding:"required"`
}

func main() {
	r := gin.Default()
	r.POST("/webhook/", func(c *gin.Context) {
		build(c)
	})
	r.Run(":8123") // listen and serve on 0.0.0.0:8080
}

func build(c *gin.Context) {
	var msg Msg

	if err := c.ShouldBindJSON(&msg); err != nil {
		lib.Logger().Errorln("参数解析错误：" + err.Error())
		resError(c, err.Error())
		return
	}

	if msg.Event_name != "push" {
		resSeccess(c)
		return
	}

	if msg.Ref != "refs/heads/dev" && msg.Ref != "refs/heads/test" {
		resSeccess(c)
		return
	}

	var env string
	if msg.Ref == "refs/heads/dev" {
		env = "dev"
	} else {
		env = "test"
	}

	// var path string
	switch msg.Project_id {
	case WEBSIETAPI:
		lib.Logger().Infoln("开始更新元宇宙API")
		path := "/www/meta-api-" + env

		ExecCommand("git", []string{"reset", "--hard"}, path)
		res, st := ExecCommand("git", []string{"pull"}, path)
		if !st {
			return
		}
		lib.Logger().Infoln("新元宇宙API拉取完成" + res)

	case WEBSIET:
		lib.Logger().Infoln("开始更新元宇宙官网")
		path := "/www/meta-web-" + env

		ExecCommand("git", []string{"reset", "--hard"}, path)
		res, st := ExecCommand("git", []string{"pull"}, path)
		if !st {
			return
		}
		lib.Logger().Infoln("拉取完成" + res)
		res, st = ExecCommand("yarn", []string{}, path)
		if !st {
			return
		}
		lib.Logger().Infoln("yarn完成：" + res)

		res, st = ExecCommand("yarn", []string{"build"}, path)
		if !st {
			return
		}
		lib.Logger().Infoln("yarn build完成：" + res)
		lib.Logger().Infoln("元宇宙官网 构建完毕")

	case ADMINWEB:
		lib.Logger().Infoln("开始更新元宇宙后台管理页面")
		path := "/www/meta-admin-web/web"

		ExecCommand("git", []string{"reset", "--hard"}, path)
		res, st := ExecCommand("git", []string{"pull"}, path)
		if !st {
			return
		}
		lib.Logger().Infoln("拉取完成" + res)

		res, st = ExecCommand("yarn", []string{}, path)
		if !st {
			return
		}
		lib.Logger().Infoln("yarn完成：" + res)

		res, st = ExecCommand("yarn", []string{"build"}, path)
		if !st {
			return
		}
		lib.Logger().Infoln("yarn build完成：" + res)
		res, st = ExecCommand("docker-compose", []string{"restart"}, path+"/../env/")
		if !st {
			return
		}
		lib.Logger().Infoln("docker-compose 完毕" + res)
		lib.Logger().Infoln("台管理页面 构建完毕")
	}

	resSeccess(c)
}

func resSeccess(c *gin.Context) {
	c.JSON(200, gin.H{
		"Code": 0,
	})
}

func resError(c *gin.Context, msg string) {
	c.JSON(500, gin.H{
		"Code": 1,
		"msg":  msg,
	})
}

func ExecCommand(command string, params []string, dir string) (string, bool) {
	cmd := exec.Command(command, params...)

	if dir != "" {
		cmd.Dir = dir
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		lib.Logger().Errorln("error   ", err)
		return "", false
	}

	return out.String(), true
}
