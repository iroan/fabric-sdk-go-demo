# fabric-sdk-go-demo
- 一个在fabric平台支持高级算法(同态计算)的实例

# 功能
1. 部署支持同态计算的链码
2. 链码的业务场景是token转账中使用同态加密技术来保证用户转账资金的隐私

# 步骤
1. 部署fabric网络
2. 部署链码
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