refactor 分支更改内容（尚未review）

**1. 身份校验完全依赖客户端传入的参数** REST 端点的 `clerk_id`、`sender_id`、`conversation_id` 全部从 URL/body 读取，而不是从已认证的 JWT session 解析。[server.go](vscode-webview://1fau0dbo58vc1e41jtbtjhqg03o6iua7f6bbvbisc5e4b9odprit/backend/cmd/api/server.go#L18) 里 Clerk 中间件的 JWT extractor 只从 `Sec-WebSocket-Protocol` 头取 token，常规 fetch 请求（[api/utils.ts](vscode-webview://1fau0dbo58vc1e41jtbtjhqg03o6iua7f6bbvbisc5e4b9odprit/frontend/api/utils.ts#L1) 没带任何认证头）根本不会被鉴权。这意味着**知道别人 clerk_id 就能伪装成他**。

- 例：[getMessages](vscode-webview://1fau0dbo58vc1e41jtbtjhqg03o6iua7f6bbvbisc5e4b9odprit/backend/cmd/api/messages.go#L43) 中鉴权代码被注释掉，任何登录用户可枚举 `conversation_id` 读取任意会话。
- **建议**：所有 REST 接口统一从 `getClerkUser(r.Context())` 解析当前用户，后端用 `users.clerk_id` 反查内部 user id，替换 URL 里的 `{clerk_id}` 参数。

### 后端（Go）

| 改动                                                                                                                       | 说明                                                                                                                                                                         |
| -------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [auth.go](vscode-webview://1fau0dbo58vc1e41jtbtjhqg03o6iua7f6bbvbisc5e4b9odprit/backend/cmd/api/auth.go)（新建）           | `requireUser` 中间件：从 Clerk JWT session claims → `GetUser(clerk_id)` 解析内部用户存入 context；`requireConversationMember`：会话成员校验（非成员 403）；`isTrustedOrigin` |
| [server.go](vscode-webview://1fau0dbo58vc1e41jtbtjhqg03o6iua7f6bbvbisc5e4b9odprit/backend/cmd/api/server.go)               | JWT extractor 合并处理 `Authorization: Bearer`（REST）+ `Sec-WebSocket-Protocol`（WS），常规请求现在也有 session claims                                                      |
| [routes.go](vscode-webview://1fau0dbo58vc1e41jtbtjhqg03o6iua7f6bbvbisc5e4b9odprit/backend/cmd/api/routes.go)               | 去掉所有 `{clerk_id}` 路径段；全部业务路由包 `requireUser`；`/health` 与 `POST /user`（webhook）保持公开                                                                     |
| [friends.go](vscode-webview://1fau0dbo58vc1e41jtbtjhqg03o6iua7f6bbvbisc5e4b9odprit/backend/cmd/api/friends.go)             | `acceptRequest` 不再接受 body 里的内部 user id，改为 `GetRequest` 后校验当前用户是 receiver；`denyRequest` 修复按错误列删除的 bug；`createRequest` sender 从 session 取      |
| [messages.go](vscode-webview://1fau0dbo58vc1e41jtbtjhqg03o6iua7f6bbvbisc5e4b9odprit/backend/cmd/api/messages.go)           | `createMessage` 忽略 body `sender_id`，用 session 用户；`markReadMessage` 忽略 `member_id`；读消息/读最后一条均加成员校验                                                    |
| [conversations.go](vscode-webview://1fau0dbo58vc1e41jtbtjhqg03o6iua7f6bbvbisc5e4b9odprit/backend/cmd/api/conversations.go) | 身份全部改为 `currentUser(r).ClerkID`；`getConversation`/`deleteFriend` 加成员校验                                                                                           |
| [ws.go](vscode-webview://1fau0dbo58vc1e41jtbtjhqg03o6iua7f6bbvbisc5e4b9odprit/backend/cmd/api/ws.go)                       | 假 `hasAccessToConversation` 换成真实成员校验；`readPump` 用 `client.userID` 覆盖伪造 sender；加 Origin 校验；`Client.userID` 改 int64                                       |
| SQL + sqlc 重生成                                                                                                          | `GetRequest`、`GetConversationMember`、`GetMessageConversationId` 新增；`DeleteRequest` 改为按 id 删                                                                         |

### 前端（Next.js）

- [api/utils.ts](vscode-webview://1fau0dbo58vc1e41jtbtjhqg03o6iua7f6bbvbisc5e4b9odprit/frontend/api/utils.ts) 新增 `authHeaders(token)`；三个 api 文件全部函数增加 `token` 参数、去掉 `clerk_id`；`markReadMessage` 去掉 `member_id`；删除未使用的 `createMessage`
- 11 个组件/hook 改为在调用点 `await getToken()`，移除 `clerkId`/`senderId`/`receiverId` props 和 `?clerk_id=` 导航链接

### 验证结果

- **后端**：`sqlc generate` ✅、`go build ./...` ✅、`go vet ./...` ✅、`gofmt` 无差异 ✅
- **前端**：`next build` webpack 编译通过 ✅；tsc 逐条对比证明**零新增类型错误**（还消除了 1 个）

### ⚠️ 两点环境情况（与本次改动无关，需你知晓）

1. **前端构建被既有依赖问题阻塞**：`next build` 在 `Collecting page data` 阶段因 React 版本不一致（`react 19.0.0` vs `react-dom 19.2.3`）失败；`npm ci` 也因 @clerk/nextjs 的 peer 冲突需 `--legacy-peer-deps`。建议后续统一 `react`/`react-dom` 版本。
2. **`CallRoom`/`UploadFileDialog` 走的是 Convex 的 `api.message.create`**（旧版残留），本次未处理——音视频通话/文件上传消息链路实际上依赖 Convex 后端。

### 下一步验证（需要你的环境）

本地跑起 postgres + backend + frontend（需要 Clerk key），按计划里的 curl/WS 冒烟用例验证：无 token → 401、伪造他人 `sender_id` → 落库为会话用户、非成员访问 → 403。需要我帮你写一个可执行的冒烟测试脚本吗？
