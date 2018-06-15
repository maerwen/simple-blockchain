package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/davecgh/go-spew/spew"

	"github.com/gorilla/mux"
)

// http://blog.hubwiz.com/2018/02/04/blockchain-diy-go/
// https://github.com/mycoralhealth/blockchain-tutorial/blob/master/networking/main.go
// 主程序
func main() {
	// 利用第三方包读取．ｅｎｖ配置文件
	// 默认读取文件为＂．ｅｎｖ＂这一个文件，可接收一个变长数组
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	// 创世块构建
	go func() {
		t := time.Now()
		genesisBlock := Block{0, t.String(), 0, "", ""}
		spew.Dump(genesisBlock)
		Blockchain = append(Blockchain, genesisBlock)
	}()
	log.Fatal(run())
}

// 定义区块链结构体
type Block struct {
	Index     int    //这个块在整个链中的位置
	Timestamp string //块生成时的时间戳
	BPM       int    // 每分钟心跳数，也就是心率
	Hash      string //块通过 SHA256 算法生成的散列值
	PrevHash  string //前一个块的 SHA256 散列值
}

// 定义POST创建链条的请求结构体
type Message struct {
	BPM int
}

//定义一个链的结构
var Blockchain []Block

//计算SHA256散列值
func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//生成块
func generateBlock(lastBlock Block, BPM int) Block {
	var newBlock Block
	t := time.Now()
	newBlock.Index = lastBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = lastBlock.Hash
	newBlock.Hash = calculateHash(newBlock)
	return newBlock
}

//校验块
func checkBlock(lastBlock Block, newBlock Block) bool {
	if lastBlock.Index != newBlock.Index-1 {
		return false
	}
	if lastBlock.Hash != newBlock.PrevHash {
		return false
	}
	if newBlock.Hash != calculateHash(newBlock) {
		return false
	}
	return true
}

//本地链条更新
func replaceChain(newChain []Block) {
	if len(newChain) > len(Blockchain) {
		Blockchain = newChain
	}
}

//web服务构建
func run() error {
	mux := makeMuxRouter()
	address := os.Getenv("PORT")
	server := &http.Server{
		Addr:           ":" + address,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

// 路由创建
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleReadBlockChain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlockChain).Methods("POST")
	return muxRouter
}

//GET查看链条
func handleReadBlockChain(w http.ResponseWriter, r *http.Request) {
	result, err := json.MarshalIndent(Blockchain, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(result)
	// io.WriteString(w, string(result))
}

//POST构建链条
func handleWriteBlockChain(w http.ResponseWriter, r *http.Request) {
	var m Message
	// 解码读取请求实体内容
	decoder := json.NewDecoder(r.Body)
	// Decode此方法所需参数必须为指针，否则报ｎｉｌ错误
	err := decoder.Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	// 根据请求实体中提供的参数与已有BlockChain生成新的Ｂｌｏｃｋ
	newBlock := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)
	// 对新生成的Ｂｌｏｃｋ进行验证，验证通过后进行链更新
	if checkBlock(Blockchain[len(Blockchain)-1], newBlock) {
		newBlockChain := append(Blockchain, newBlock)
		replaceChain(newBlockChain)
		//将struct、slice格式化输出到控制台
		spew.Dump(Blockchain)
	}
	// 更新后将整个链条内容发送到前端页面
	handleReadBlockChain(w, r)
}
