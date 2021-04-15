这是一个在fabric平台支持高级算法(同态计算)的示例

# 功能
1. 部署支持同态计算的链码
2. 链码的业务场景是token转账中使用同态加密技术来保证用户转账资金的隐私

# 步骤
1. 部署fabric网络、链码（需要确定环境是否一致）: `./start-node-and-deploy-cc.sh`
3. 用户注册
4. 用户转账

# 演示
```zsh
➜  fabric-sdk-go-demo git:(main) ✗ go test -v -run  ^TestTokenTransfer$            
=== RUN   TestTokenTransfer
200 Admin@org1.example.com enroll successfully
200 User1@org1.example.com enroll successfully
|admin balance       |user1 balance       |
|100                 |100                 |
|95                  |105                 |
|103                 |97                  |
--- PASS: TestTokenTransfer (12.79s)
PASS
ok      github.com/iroan/fabric-sdk-go-demo     12.839s
```

# 环境
```
➜  ~ peer version
peer:
 Version: 2.2.3
 Commit SHA: 496c5f547
 Go version: go1.16.2
 OS/Arch: linux/amd64
 Chaincode:
  Base Docker Label: org.hyperledger.fabric
  Docker Namespace: hyperledger
```

另外项目fabric的Go SDK(`fabric-sdk-go`)中
1. 官方测试代码并不会去读取环境变量`CRYPTOCONFIG_FIXTURES_PATH`的值
2. 其直接返回硬编码的变量`CryptoConfigPath`的值，[link](https://github.com/hyperledger/fabric-sdk-go/blob/c4d51626e6c9fc82432f76c20b01d0dfa709b22a/test/metadata/metadata.go#L14)
3. 因为[crypto等文件的路径](https://github.com/hyperledger/fabric-sdk-go/blob/c4d51626e6c9fc82432f76c20b01d0dfa709b22a/test/fixtures/config/config_e2e.yaml#L67)是拼接出来的，最好使用相对路径
4. 如果crypto路径不在`FABRIC_SDK_GO_PROJECT_PATH`的路径下，可以软链接到该目录下(这里链接到仓库根目录的organizations)
