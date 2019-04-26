# hashgardcli issue burn-any-off

## 描述
Owner关闭自己发行的某个代币的可销毁任意持币者的该币的功能。

>注：一旦关闭便不可再打开。

## 使用方式
```bash
 hashgardcli issue burn-any-off [issue-id] [flags]
```
## Global Flags

 ### 参考：[hashgardcli](../README.md)

## 示例
### 关闭销毁功能
```shell
hashgardcli issue burn-any-off coin174876e800 --from foo
```
输入正确的密码之后，你的该代币便关闭了可由Owner销毁的功能。
```json
{
 "height": "4917",
 "txhash": "B9F97B17BD986E9FA7CF41EF2FDF844E8D9582987D5183FB160A0FBDD6A7B045",
 "data": "ERBjb2luMTU1NTU2NzUwNjAw",
 "logs": [
  {
   "msg_index": "0",
   "success": true,
   "log": ""
  }
 ],
 "gas_wanted": "100000000",
 "gas_used": "9086502",
 "tags": [
  {
   "key": "action",
   "value": "issue_burn_any_off"
  },
  {
   "key": "issue-id",
   "value": "coin174876e800"
  }
 ]
}
```