# 方案简述
本仓库开源的是一种基于区块链的人工智能生成内容多媒体数字水印方法，通过预处理区块链中的Nonce值，优化交易流程，上链过程与水印嵌入解耦又能相互关联，有效解决区块链和AIGC的版权应用融合落地难题。适用于视频、音频、图片等其他多媒体AIGC, 上链记录AIGC行为中的训练痕迹和用户干预行为，在对产出内容嵌入更好鲁棒性的数字水印同时，进行隐私加密，实现合规监测、溯源追踪、侵权防范等功能。

## 该仓库为本方案的Demo实现，步骤如下：
## 下载stable diffusion
我在本地搭建SD来作为方案中的AIGC平台，stable diffusion的安装流程这里不赘述了，大家可自行去搜索教程搭建或是使用其他的GC平台。

## 修改xuper源码
### 部署xuper并跑通示例
我使用的 `V5.3` 版本部署。  
文档地址链接：[Xuper文档地址][https://xuper.baidu.com/n/xuperdoc/v5.3/quickstart/quickstart.html]

### 源码修改
通过修改Xuperchain底层的Nonce处理逻辑，来达到使用嵌入到水印中的Nonce值作为交易Nonce的目的。  
修改处为：kernel\engines\xuperos\chain.go 中的 SubmitTx函数，详情可参考仓库中的chain.go文件。可通过两个方法来修改源码：  
`方法一` ：
查看xuperchain目录下的go.mod文件，找到 github.com/xuperchain/xupercore 对应的版本如 v0.0.0-20221206131501-5a3396e9215d

前往你的GOPATH目录所在，可通过 go env 查看。  
cd $GOPATH/pkg/mod/github.com/xuperchain/

使用仓库中的chain.go替换 xupercore@v0.0.0-20221206131501-5a3396e9215d 目录下的chain.go

停止xuperchain示例网络，再次编译启动，命令如下：
///Bash
bash control.sh stop
cd ..
make all
bash control.sh start
///
