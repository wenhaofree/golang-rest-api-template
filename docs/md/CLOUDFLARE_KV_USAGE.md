# Cloudflare KV 使用说明（tests/kv.go）

本文档说明如何使用本仓库提供的测试程序 `tests/kv.go` 通过 Cloudflare Workers KV 的 REST API 进行写入与读取验证。

## 功能概述
- 通过 REST API 对 KV 执行单 Key 写入（PUT）与读取（GET）。
- 支持为写入设置可选的过期参数：`expiration_ttl`（相对过期，单位秒）或 `expiration`（绝对过期，UNIX 时间戳秒）。
- 通过环境变量传递账号、命名空间、鉴权 Token 与测试 Key/Value。

> 说明：Cloudflare 官方建议外部应用使用 REST API 访问 Workers KV，PUT/GET 的端点格式如下：`/accounts/{account_id}/storage/kv/namespaces/{namespace_id}/values/{key_name}`（鉴权使用 `Authorization: Bearer <API_TOKEN>`）。参考：
> - Workers KV 概览与 REST API： https://developers.cloudflare.com/kv/
> - 写入单个 Key（REST API 文档）： https://developers.cloudflare.com/api/resources/kv/subresources/namespaces/subresources/values/methods/update/
> - 读取 Key（REST API 或绑定 API 说明）： https://developers.cloudflare.com/kv/api/read-key-value-pairs/

## 先决条件
1. 你已创建 Cloudflare 账号与目标 KV Namespace（可在 Cloudflare Dashboard 上创建）。
2. 你已创建一个具备 Workers KV Storage 读写权限的 API Token，并获得以下信息：
   - `CF_API_TOKEN`：API Token（具备 KV 读写权限）
   - `CF_ACCOUNT_ID`：Cloudflare 账号 ID
   - `CF_KV_NAMESPACE_ID`：目标 KV 命名空间 ID

## 环境变量
测试程序从环境变量中读取配置：

必填：
- `CF_API_TOKEN`：Cloudflare API Token（需具备 Workers KV Storage 的读写权限）
- `CF_ACCOUNT_ID`：Cloudflare Account ID
- `CF_KV_NAMESPACE_ID`：KV Namespace ID

可选：
- `CF_KV_TEST_KEY`：测试用的 Key，默认 `golang-rest-api-template:test`
- `CF_KV_TEST_VALUE`：测试写入的值，默认 `hello-from-go`
- `CF_KV_EXPIRATION_TTL`：相对过期时间（单位秒，最小 60），设置后优先生效
- `CF_KV_EXPIRATION`：绝对过期时间（UNIX epoch 秒），当未设置 TTL 时生效
- `CF_KV_BASE_URL`：Cloudflare API 基础地址，默认 `https://api.cloudflare.com/client/v4`

## 运行方式
在仓库根目录执行：

```bash
# 确保已设置必要环境变量
export CF_API_TOKEN=xxxxx
export CF_ACCOUNT_ID=yyyyy
export CF_KV_NAMESPACE_ID=zzzzz

# 可选环境变量
export CF_KV_TEST_KEY="golang-rest-api-template:test"
export CF_KV_TEST_VALUE="hello-from-go"
# 注意：TTL 最小为 60 秒，若设置则优先于 expiration 生效
# export CF_KV_EXPIRATION_TTL=120
# export CF_KV_EXPIRATION=1735689600

# 运行测试程序
go run ./tests/kv.go
```

程序流程：
1. 校验必填环境变量。
2. 调用 PUT 写入 Key/Value（附带可选的过期参数）。
3. 稍作延迟后调用 GET 读取 Key。
4. 比较读回值与写入值，成功则输出 `KV write/read succeeded.`。

## 常见问题与排错
- 403/401 鉴权失败：
  - 检查 `CF_API_TOKEN` 是否正确，且具备 Workers KV Storage 写/读权限。
  - 检查是否使用了 `Authorization: Bearer <token>` 的 Bearer Token 鉴权方式。
- 404 未找到：
  - 检查 `CF_ACCOUNT_ID`、`CF_KV_NAMESPACE_ID` 是否正确。
  - 检查 `CF_KV_TEST_KEY` 是否拼写一致。
- 429 速率限制：
  - 对同一 Key，Workers KV 限制写入速率（通常为 1 次/秒）。若出现 429，建议在上层增加重试与指数退避。
- 最终一致性与传播延迟：
  - KV 的写入为最终一致，跨全球边缘位置传播可能需要一定时间（通常可达 60 秒）。本测试仅做了极短延迟；若在你的场景中出现刚写入就读取不到的情况，可适当延长读取前的延迟或增加重试策略。

## 扩展与参考
- 如需写入带 metadata 的值，REST API 支持通过 `multipart/form-data` 携带 `metadata` 字段；当前 `tests/kv.go` 未实现，可在需要时扩展。
- 批量写入（bulk）也可通过 REST API 完成，适合一次写入大量 KV 对（单次请求上限与体积限制详见官方文档）。

参考文档：
- Cloudflare Workers KV（概览与入口）：https://developers.cloudflare.com/kv/
- 写入单个 Key（REST API）：https://developers.cloudflare.com/api/resources/kv/subresources/namespaces/subresources/values/methods/update/
- 读取 Key（绑定/REST 说明）：https://developers.cloudflare.com/kv/api/read-key-value-pairs/

## 安全建议
- 切勿将 `CF_API_TOKEN` 等敏感凭据提交到版本库。
- 建议使用本地环境变量、密钥管理服务或 CI/CD 的安全变量存储 Token。