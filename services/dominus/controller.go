package dominus

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sdslabs/gasper/lib/middlewares"
	"github.com/sdslabs/gasper/lib/mongo"
	"github.com/sdslabs/gasper/lib/redis"
	"github.com/sdslabs/gasper/types"
)

func createApp(c *gin.Context) {
	instanceURL, err := redis.GetLeastLoadedWorker()
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if instanceURL == redis.ErrEmptySet {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "No worker instances available at the moment",
		})
		return
	}
	reverseProxy(c, instanceURL)
}

func createDatabase(c *gin.Context) {
	database := c.Param("database")
	instanceURL, err := redis.GetLeastLoadedInstance(database)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if instanceURL == redis.ErrEmptySet {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "No worker instances available at the moment",
		})
		return
	}
	reverseProxy(c, instanceURL)
}

func execute(c *gin.Context) {
	app := c.Param("app")
	instanceURL, err := redis.FetchAppNode(app)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Application %s is not deployed at the moment", app),
		})
		return
	}
	reverseProxy(c, instanceURL)
}

func deleteDB(c *gin.Context) {
	db := c.Param("db")
	instanceURL, err := redis.FetchDBURL(db)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "No such database exists",
		})
		return
	}
	reverseProxy(c, instanceURL)
}

func fetchInstancesByUser(instanceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userStr := middlewares.ExtractClaims(c)
		filter := types.M{
			"instanceType": instanceType,
			"owner":        userStr.Email,
		}
		c.JSON(200, gin.H{
			"success": true,
			"data":    mongo.FetchAppInfo(filter),
		})
	}
}

func fetchAppsByUser() gin.HandlerFunc {
	return fetchInstancesByUser(mongo.AppInstance)
}

func fetchDbsByUser() gin.HandlerFunc {
	return fetchInstancesByUser(mongo.DBInstance)
}
