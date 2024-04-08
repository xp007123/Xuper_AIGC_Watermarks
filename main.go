package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
)

// Config 存储应用程序的配置信息
type Config struct {
	RedisAddr       string
	XuperchainAddr  string
	ContractAccount string
}

// Content 表示要存储在Redis中的数据结构
type Content struct {
	Nonce string `json:"nonce"`
	Infor string `json:"infor"`
	Time  int64  `json:"time"`
}

var (
	ctx     = context.Background() // 上下文对象,用于Redis操作
	config  = loadConfig()         // 加载应用程序配置
	redismu sync.Mutex             // 互斥锁,用于保护Redis连接
	rdb     *redis.Client          // Redis客户端连接
)

// loadConfig 从外部源加载应用程序配置
func loadConfig() Config {
	// 从配置文件或其他来源加载配置
	return Config{
		RedisAddr:       "192.168.56.101:6379",  // Redis地址
		XuperchainAddr:  "192.168.56.101:37101", // Xuperchain节点地址
		ContractAccount: "XC7777777777777777@xuper",
	}
}

// genPseudoUniqID 生成伪唯一ID，用做Nonce
func genPseudoUniqID() uint64 {
	nano := time.Now().UnixNano()

	randNum1 := rand.Int63()
	randNum2 := rand.Int63()
	shift1 := rand.Intn(16) + 2
	shift2 := rand.Intn(8) + 1

	uID := ((randNum1 >> uint(shift1)) + (randNum2 >> uint(shift2)) + (nano >> 1)) &
		0x1FFFFFFFFFFFFF
	return uint64(uID)
}

// getRedisClient 获取Redis客户端连接,确保并发安全
func getRedisClient() *redis.Client {
	redismu.Lock()
	defer redismu.Unlock()

	if rdb == nil {
		rdb = redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr, // 使用配置中的Redis地址
			Password: "",               // 没有设置密码
			DB:       0,                // 使用默认数据库
		})
	}

	return rdb
}

// xuperchain 调用Xuperchain智能合约,存储数据
func xuperchain(nonce, infor string) (string, error) {
	client, err := xuper.New(config.XuperchainAddr) // 使用配置中的Xuperchain节点地址
	if err != nil {
		return "", err
	}
	defer client.Close() // 在函数结束时关闭Xuperchain客户端连接

	acc, err := account.GetAccountFromPlainFile("./cert/keys") // 从文件加载账户信息
	if err != nil {
		return "", err
	}

	err = acc.SetContractAccount(config.ContractAccount) // 设置合约账户
	if err != nil {
		return "", err
	}

	contractName := "watermarkasss" // 合约名称
	contractMethod := "storeInfo"   // 要调用的合约方法
	args := map[string]string{      // 合约方法参数
		"nonce":     nonce,
		"otherInfo": infor,
	}

	tx, err := client.InvokeEVMContract(acc, contractName, contractMethod, args) // 调用合约
	if err != nil {
		return "", err
	}

	txid := hex.EncodeToString(tx.Tx.Txid) // 获取交易ID
	txQu, _ := client.QueryTxByID(txid)    // 查询交易信息
	fmt.Printf("Transaction ID: %s\n", txid)
	fmt.Printf("Nonce: %s\n", txQu.GetNonce())

	return txid, nil // 返回交易ID
}

func main() {
	r := gin.Default() // 创建Gin Web服务器

	// 处理POST请求"/start"
	r.POST("/start", func(c *gin.Context) {
		nonce := strconv.Itoa(int(genPseudoUniqID())) // 生成伪唯一ID

		ctt := Content{Nonce: nonce, Infor: "", Time: time.Now().Unix()} // 创建Content结构体
		jsonData, err := json.Marshal(ctt)                               // 将Content结构体序列化为JSON
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		rdb := getRedisClient()
		err = rdb.Set(ctx, config.ContractAccount, jsonData, 0).Err() // 将JSON数据存储在Redis中
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"Nonce": nonce,
		})
	})

	// 处理POST请求"/update"
	r.POST("/update", func(c *gin.Context) {
		sdInfor := c.PostForm("sdInfor") // 获取POST表单中的"sdInfor"字段

		rdb := getRedisClient()
		val, err := rdb.Get(ctx, config.ContractAccount).Result() // 从Redis中获取config.ContractAccount的值
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		var data map[string]interface{}
		err = json.Unmarshal([]byte(val), &data) // 将JSON数据反序列化为map
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		data["infor"] = data["infor"].(string) + sdInfor // 更新"infor"字段
		data["time"] = time.Now().Unix()                 // 更新"time"字段
		newJson, err := json.Marshal(data)               // 将更新后的数据序列化为JSON
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		err = rdb.Set(ctx, config.ContractAccount, newJson, 0).Err() // 将更新后的JSON数据存储在Redis中
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Update successful",
		})
	})

	r.POST("/end", func(c *gin.Context) {
		rdb := getRedisClient()
		val, err := rdb.Get(ctx, config.ContractAccount).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		var data map[string]interface{}
		err = json.Unmarshal([]byte(val), &data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		nonce := data["nonce"].(string)
		infor := data["infor"].(string)
		txid, err := xuperchain(nonce, infor)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"txid": txid,
		})
	})

	r.Run()
}
