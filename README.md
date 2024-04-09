# 方案简述
本仓库开源的是一种基于区块链的人工智能生成内容多媒体数字水印方法，通过预处理区块链中的Nonce值，优化交易流程，上链过程与水印嵌入解耦又能相互关联，有效解决区块链和AIGC的版权应用融合落地难题。适用于视频、音频、图片等其他多媒体AIGC, 上链记录AIGC行为中的训练痕迹和用户干预行为，在对产出内容嵌入更好鲁棒性的数字水印同时，进行隐私加密，实现合规监测、溯源追踪、侵权防范等功能。

## 该仓库为本方案的Demo实现，步骤如下：
## 下载stable diffusion
- 我在本地搭建SD来作为方案中的AIGC平台，stable diffusion的安装流程这里不赘述了，大家可自行去搜索教程搭建或是使用其他的GC平台。

## 修改Stable Diffusion源码
修改内容  | 修改文件
------------- | -------------
添加生成图片后水印处理  | stable-diffusion-webui-master\modules\images.py
添加生成图片时与XuperChain的交互处理  | stable-diffusion-webui-master\modules\images.py
在前端页面中添加上链按钮及反馈  | stable-diffusion-webui-master\modules\ui.py

- 以上两个修改文件都在放在了仓库之中，可直接替换，在文件中搜索`amend`可定位查询修改的地方。
  
- 前端页面的修改效果如图所示：
![_98OUV}7FN3G@38DA~3 Z(C](https://github.com/xp007123/Xuper_AIGC_Watermarks/assets/57866608/a0bed4d0-1bfe-4b94-b57f-8ad68a193338)

## 修改xuper源码
### 部署xuper并跑通示例
- 我使用的 `V5.3` 版本部署。[Xuper文档地址](https://xuper.baidu.com/n/xuperdoc/v5.3/quickstart/quickstart.html)   
- 该Demo会使用redis功能，请在你的服务器上启动。   

### 源码修改
- 通过修改Xuperchain底层的Nonce处理逻辑，来达到使用嵌入到水印中的Nonce值作为交易Nonce的目的。     

- 修改处为：kernel\engines\xuperos\chain.go 中的 SubmitTx函数，详情可参考仓库中的chain.go文件。可通过两个方法来修改源码：  

`方法一` 
- 查看xuperchain目录下的go.mod文件，找到 github.com/xuperchain/xupercore 对应的版本如 v0.0.0-20221206131501-5a3396e9215d

- 前往你的GOPATH目录所在，可通过 go env 查看。
```Bash 
cd $GOPATH/pkg/mod/github.com/xuperchain/
```

- 使用仓库中的chain.go替换 xupercore@v0.0.0-20221206131501-5a3396e9215d 目录下的chain.go

- 停止xuperchain示例网络，再次编译启动，命令如下：
```Bash
bash control.sh stop  
cd ..  
make all  
bash control.sh start  
```
`方法二` 
- 不直接修改Xuperchain的底层包
```Bash
cd $GOPATH/pkg/mod/github.com/xuperchain/
cp -r xupercore@v0.0.0-20221206131501-5a3396e9215d xupercore@v0.0.0-amend
```  
- 使用仓库中的chain.go替换 xupercore@v0.0.0-amend 目录下的chain.go

- 在xuperchain目录下的go.mod文件中添加replace命令
```Bash
vim go.mod
replace github.com/xuperchain/xupercore => /home/chalken/GOPATH/pkg/mod/github.com/xuperchain/xupercore@v0.0.0-amend-20221206131501
```
- 在退出go.mod后，停止停止xuperchain示例网络，make后再次启动
```Bash
go mod tidy
make all
```
## 部署Demo合约
- 合约文件是仓库中的 `Watermarkasss.sol`, 可将该文件直接放置在output目录下，部署命令如下：
```Bash
bin/xchain-cli account new --account 7777777777777777 --fee 2000
bin/xchain-cli transfer --to XC7777777777777777@xuper --amount 100000000 --keys data/keys/ -H 127.0.0.1:37101
solc --bin --abi Watermarks.sol -o .
bin/xchain-cli evm deploy --account XC7777777777777777@xuper --cname watermarkasss  --fee 5200000 Watermarkasss.bin --abi Watermarkasss.abi
```

## 启动XuperChain服务端（xuper-sdk-go）
- 我用的是XuperChain的golang SDK版本。[XuperGolangSDK文档地址](https://xuper.baidu.com/n/xuperdoc/v5.3/development_manuals/xuper-sdk/xuper-sdk-go.html)

- 在SDK连接你的XuperChain测试网络无误后，使用仓库中的main.go文件启动服务端，请注意修改main.go文件中的redis及节点IP等配置。

## 效果验证
- 在你启动了Stable Diffusion 和 XuperChain服务端后，请在Stable Diffusion页面输入一组Prompt提示词，并调整Sampling steps为5 （生成快一些）。

- 图片生成之后，我们可以在 stable-diffusion-webui-master\outputs\txt2img-images 目录下看到生成的图片，其中 `XXX_3we.png` 图片为添加了水印的图片。
  
- 在生成几张图片后点击 `Copyright Protection` 按钮，该按钮会使用嵌入在水印中的nonce值将Prompt，图片生成设置等信息全部上链。

- 我们在[回到顶部](#readme)的图片中的可以看到上链完成会返回txid：66d20173951c8b883f12df720d452d919284b91f154ba991af4ea3716a9d833b

- 前往XuperChain的网络环境下查询交易：
```Bash
bin/xchain-cli tx query 66d20173951c8b883f12df720d452d919284b91f154ba991af4ea3716a9d833b
```

- 我们可以看到交易中包含了nonce值和用户的Prompt，图片生成设置等信息，且nonce值与交易头中的nonce值一致
![F}1T00(S$3`1RX1O$~)014I](https://github.com/xp007123/Xuper_AIGC_Watermarks/assets/57866608/5e634ccc-8718-4d33-9f27-a2b946b7d39b)
![OQHOD7TQGZKSXOR%GC3{I52](https://github.com/xp007123/Xuper_AIGC_Watermarks/assets/57866608/f7835e59-3559-4916-ab91-2ede2470c82b)

- 此时我们提取水印看下刚刚所生成的图片中的水印值是否都是该交易的nonce值
- 运行下面的python程序，注意修改路径为你的图片路径。我们可以看到提取到的nonce值与交易中的nonce一致。
```python
from blind_watermark import WaterMark

bwm1 = WaterMark(password_img=1, password_wm=1)
wm_extract = bwm1.extract("D:\\code\\stable-diffusion-webui-master\\outputs\\txt2img-images\\2024-04-08\\00007-2479572804_3we.png", 
                          wm_shape=126, mode='str')
print(wm_extract)
```
![CJ~Z4052AY1(TODJW1LT3X2](https://github.com/xp007123/Xuper_AIGC_Watermarks/assets/57866608/2ce6a144-b7cc-43cc-8c07-0d75d7029a60)

**方案Demo演示到此结束，该Demo是方案的核心逻辑，实现了Nonce的嵌入并与链上对应起来就可基于该特征拓展实现诸多功能，解决AIGC场景下全部作品上链版权存证和监管问题，大幅度减少了交易开销，提高了区块链+AIGC版权应用大规模落地的可能性**
**
