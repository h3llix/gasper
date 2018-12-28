package php

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sdslabs/SWS/lib/api"
	"github.com/sdslabs/SWS/lib/configs"
	"github.com/sdslabs/SWS/lib/mongo"
	"github.com/sdslabs/SWS/lib/types"
	"github.com/sdslabs/SWS/lib/utils"
)

// createApp function handles requests for making making new php app
func createApp(c *gin.Context) {
	var (
		data map[string]interface{}
	)

	c.BindJSON(&data)

	ports, err := utils.GetFreePorts(2)

	if err != nil {
		c.JSON(200, gin.H{
			"error": err,
		})
		return
	}

	if len(ports) < 2 {
		c.JSON(200, gin.H{
			"error": "Not Enough Ports",
		})
		return
	}

	sshPort := ports[0]
	httpPort := ports[1]

	appEnv, err := api.CreateBasicApplication(
		data["name"].(string),
		data["location"].(string),
		data["url"].(string),
		strconv.Itoa(httpPort),
		strconv.Itoa(sshPort),
		&types.ApplicationConfig{
			DockerImage:  "nginx",
			ConfFunction: configs.CreateStaticContainerConfig,
		})
	// A hack for the nil != nil problem ( Comparing interface with a true nil value )
	var check *types.ResponseError
	if err != check {
		c.JSON(200, gin.H{
			"error": err,
		})
		return
	}

	composerPath := data["composerPath"].(string)

	// Perform composer install in the container
	if data["composer"].(bool) == true {
		execID, err := installPackages(composerPath, appEnv)
		if err != check {
			c.JSON(200, gin.H{
				"error": err,
			})
			return
		}
		data["execID"] = execID
	}

	data["sshPort"] = sshPort
	data["httpPort"] = httpPort
	data["containerID"] = appEnv.ContainerID
	data["language"] = "php"
	data["hostIP"] = utils.HostIP

	documentID, err := mongo.RegisterApp(data)

	if err != nil {
		c.JSON(200, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"id":      documentID,
	})
}

func fetchDocs(c *gin.Context) {
	queries := c.Request.URL.Query()
	filter := utils.QueryToFilter(queries)

	filter["language"] = "php"

	c.JSON(200, gin.H{
		"data": mongo.FetchAppInfo(filter),
	})
}

func deleteApp(c *gin.Context) {
	queries := c.Request.URL.Query()
	filter := utils.QueryToFilter(queries)

	filter["language"] = "php"

	c.JSON(200, gin.H{
		"message": mongo.DeleteApp(filter),
	})
}

func updateApp(c *gin.Context) {
	queries := c.Request.URL.Query()
	filter := utils.QueryToFilter(queries)

	filter["language"] = "php"

	var (
		data map[string]interface{}
	)
	c.BindJSON(&data)

	c.JSON(200, gin.H{
		"message": mongo.UpdateApp(filter, data),
	})
}
