# doh_whitelist 插件使用文档

## 插件简介

`doh_whitelist` 是一个用于限制 DoH (DNS over HTTPS) 请求访问的插件，支持基于 IP 地址和 URL 路径的双重白名单鉴权机制。

## 功能特性

- ✅ **仅对 DoH 请求生效**：自动识别 HTTPS/H2/H3 协议，不影响其他 DNS 协议
- ✅ **IP 地址白名单**：支持单个 IP 和 CIDR 网段
- ✅ **URL 路径白名单**：通过路径 token 实现鉴权
- ✅ **路径自动 ECS**：为不同路径自动添加不同的 ECS IP（运营商分流），并将配置过 `path_ecs` 的路径视为白名单路径
- ✅ **灵活匹配模式**：支持 OR（任一匹配）或 AND（同时匹配）逻辑
- ✅ **动态加载**：支持从数据提供者动态加载 IP 白名单
- ✅ **可配置响应码**：自定义拒绝请求时的 DNS 响应码
- ✅ **智能 IP 选择**：根据查询类型（A/AAAA）自动选择 IPv4 或 IPv6

## 配置参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `whitelist` | `[]string` | 否 | IP 地址白名单列表 |
| `path_list` | `[]string` | 否 | URL 路径白名单列表 |
| `path_ecs` | `map[string]PathECS` | 否 | 路径到 ECS IP 的映射配置 |
| `require_both` | `bool` | 否 | 是否要求 IP 和路径同时匹配（默认：false） |
| `rcode` | `int` | 否 | 拒绝时的 DNS 响应码（默认：5 = REFUSED） |

**PathECS 配置项：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `ipv4` | `string` | 否 | IPv4 地址（用于 ECS） |
| `ipv6` | `string` | 否 | IPv6 地址（用于 ECS） |
| `mask4` | `int` | 否 | IPv4 子网掩码（默认：24） |
| `mask6` | `int` | 否 | IPv6 子网掩码（默认：48） |

**注意**：`whitelist`、`path_list` 和 `path_ecs` 至少需要配置一个。

## 配置示例

### 示例 1：仅使用 IP 白名单

适用于内网环境，只允许特定 IP 访问 DoH 服务。

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      whitelist:
        - "192.168.1.0/24"      # 允许整个内网段
        - "10.0.0.1"            # 允许特定 IP
        - "172.16.0.0/16"       # 允许更大的内网段
        - "2001:db8::/32"       # 允许 IPv6 网段
```

**使用场景：**
- 内网 DNS 服务器，只允许内网客户端访问
- 限制 DoH 服务仅对特定 IP 范围开放

### 示例 2：仅使用路径白名单（推荐用于公网服务）

适用于公网环境，通过 URL 路径 token 实现鉴权。

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      path_list:
        - "/dns-query/token123"
        - "/dns-query/secret-key-456"
        - "/dns-query/my-custom-path"
        - "/custom-dns/another-token"
```

**客户端访问示例：**
```bash
# 使用 curl 测试
curl -X POST "https://your-server.com/dns-query/token123" \
  -H "Content-Type: application/dns-message" \
  --data-binary @query.bin

# 使用 dig 测试（需要支持 DoH 的客户端）
dig @https://your-server.com/dns-query/token123 example.com
```

**使用场景：**
- 公网 DoH 服务，通过 token 控制访问
- 多租户环境，不同用户使用不同路径 token
- 防止未授权访问，隐藏真实服务路径

### 示例 3：IP 或路径任一匹配（默认模式）

内网用户直接访问，公网用户需要 token。

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      whitelist:
        - "192.168.1.0/24"      # 内网 IP 白名单
        - "10.0.0.0/8"          # 更大的内网段
      path_list:
        - "/dns-query/public-token"  # 公网路径白名单
      # require_both: false  # 默认值，IP 或路径任一匹配即可
```

**工作原理：**
- 内网用户（192.168.1.0/24）：可以直接访问任何路径 ✅
- 公网用户：必须使用 `/dns-query/public-token` 路径才能访问 ✅
- 公网用户使用错误路径：拒绝访问 ❌

**使用场景：**
- 混合环境：内网用户便捷访问，公网用户需要认证
- 逐步迁移：从 IP 白名单迁移到路径 token 认证

### 示例 4：IP 和路径必须同时匹配（严格模式）

要求同时满足 IP 和路径条件。

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      whitelist:
        - "10.0.0.0/8"          # 仅允许内网
      path_list:
        - "/dns-query/admin-key" # 仅允许管理员路径
      require_both: true         # IP 和路径必须同时匹配
```

**工作原理：**
- 内网 IP + 正确路径：允许访问 ✅
- 内网 IP + 错误路径：拒绝访问 ❌
- 外网 IP + 正确路径：拒绝访问 ❌
- 外网 IP + 错误路径：拒绝访问 ❌

**使用场景：**
- 高安全要求：需要同时满足 IP 和路径条件
- 多级认证：IP 认证 + 路径 token 双重验证

### 示例 5：从数据提供者加载 IP 白名单

支持从数据提供者动态加载 IP 列表。

```yaml
data_providers:
  - tag: trusted_ips
    type: file
    args:
      path: ./trusted_ips.txt

plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      whitelist:
        - "192.168.1.0/24"
        - "provider:trusted_ips"  # 从数据提供者加载
      path_list:
        - "/dns-query/token123"
```

**数据提供者文件格式（trusted_ips.txt）：**
```
# 注释行
10.0.0.1
10.0.0.2
172.16.0.0/16
2001:db8::/32
```

### 示例 6：路径自动添加 ECS（运营商分流）

根据不同的路径自动添加不同运营商的 ECS IP，实现运营商分流：

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      path_ecs:
        # 电信路径：使用电信 IP 作为 ECS
        "/dns-query/telecom":
          ipv4: "1.2.3.4"      # 电信 IP
          ipv6: "2001:db8::1"  # 电信 IPv6（可选）
          mask4: 24
          mask6: 48
        
        # 移动路径：使用移动 IP 作为 ECS
        "/dns-query/mobile":
          ipv4: "5.6.7.8"      # 移动 IP
          mask4: 24
        
        # 联通路径：使用联通 IP 作为 ECS
        "/dns-query/unicom":
          ipv4: "9.10.11.12"   # 联通 IP
          mask4: 24
        
        # 广电路径：使用广电 IP 作为 ECS
        "/dns-query/cable":
          ipv4: "13.14.15.16"  # 广电 IP
          mask4: 24
```

**工作原理：**
- 客户端访问 `/dns-query/telecom` → 自动添加电信 IP 的 ECS → 上游 DNS 返回电信优化的结果
- 客户端访问 `/dns-query/mobile` → 自动添加移动 IP 的 ECS → 上游 DNS 返回移动优化的结果
- 客户端访问 `/dns-query/unicom` → 自动添加联通 IP 的 ECS → 上游 DNS 返回联通优化的结果

**使用场景：**
- CDN 优化：不同运营商用户使用不同路径，获得对应运营商的最优节点
- 多线路 DNS：根据路径自动选择对应的运营商线路
- 负载均衡：通过路径分流到不同的上游 DNS 服务器

### 示例 7：路径白名单 + 自动 ECS

结合路径白名单和自动 ECS：

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      # 路径白名单（允许访问的路径）
      path_list:
        - "/dns-query/telecom"
        - "/dns-query/mobile"
        - "/dns-query/unicom"
      
      # 路径 ECS 配置（自动添加 ECS）
      path_ecs:
        "/dns-query/telecom":
          ipv4: "1.2.3.4"
          mask4: 24
        "/dns-query/mobile":
          ipv4: "5.6.7.8"
          mask4: 24
        "/dns-query/unicom":
          ipv4: "9.10.11.12"
          mask4: 24
```

### 示例 8：自定义拒绝响应码

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      path_list:
        - "/dns-query/token123"
      rcode: 2  # 使用 SERVFAIL 而不是默认的 REFUSED
```

**常用响应码：**
- `0` = NOERROR（不推荐，可能泄露信息）
- `2` = SERVFAIL（服务器错误）
- `3` = NXDOMAIN（域名不存在）
- `5` = REFUSED（拒绝，默认值）

## 在插件链中使用

### 基本用法

将插件放在处理链的前端，尽早过滤未授权请求：

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      path_list:
        - "/dns-query/token123"

  - tag: main_sequence
    type: sequence
    args:
      exec:
        - doh_whitelist  # 首先检查白名单
        - cache
        - forward
```

### 配合其他插件使用

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      whitelist:
        - "192.168.1.0/24"
      path_list:
        - "/dns-query/token123"

  - tag: client_limiter
    type: client_limiter
    args:
      max_qps: 100

  - tag: main_sequence
    type: sequence
    args:
      exec:
        - doh_whitelist      # 1. 白名单检查
        - client_limiter     # 2. 限流
        - cache              # 3. 缓存
        - forward            # 4. 转发
```

## 路径匹配规则

### 路径规范化

插件会自动规范化路径：
- 自动添加前导斜杠：`dns-query/token` → `/dns-query/token`
- 移除尾随斜杠：`/dns-query/token/` → `/dns-query/token`
- 空路径或 `/` 保持不变

### 匹配示例

| 配置路径 | 请求路径 | 是否匹配 |
|---------|---------|---------|
| `/dns-query/token` | `/dns-query/token` | ✅ |
| `/dns-query/token` | `/dns-query/token/` | ✅（自动规范化） |
| `dns-query/token` | `/dns-query/token` | ✅（自动添加斜杠） |
| `/dns-query/token` | `/dns-query/wrong` | ❌ |
| `/dns-query/token` | `/dns-query/token123` | ❌（精确匹配） |

**注意**：路径匹配是精确匹配，不支持通配符或正则表达式。

## 客户端配置示例

### 使用 curl 测试

```bash
# 准备 DNS 查询（base64 编码）
echo -n "AAABAAABAAAAAAAAB2V4YW1wbGUDY29tAAABAAE=" | base64 -d > query.bin

# 发送 DoH 请求（使用路径 token）
curl -X POST "https://your-server.com/dns-query/token123" \
  -H "Content-Type: application/dns-message" \
  -H "Accept: application/dns-message" \
  --data-binary @query.bin
```

### 使用 dnspython 测试

```python
import dns.message
import dns.query
import requests

# 创建查询
q = dns.message.make_query('example.com', 'A')

# DoH 请求（使用路径 token）
url = "https://your-server.com/dns-query/token123"
response = requests.post(
    url,
    data=q.to_wire(),
    headers={'Content-Type': 'application/dns-message'}
)

# 解析响应
r = dns.message.from_wire(response.content)
print(r)
```

### 配置客户端使用路径 token

**使用 cloudflared：**
```bash
cloudflared proxy-dns --upstream "https://your-server.com/dns-query/token123"
```

**使用 stubby：**
```yaml
upstream_recursive_servers:
  - address_data: your-server.com
    tls_port: 443
    tls_auth_name: "your-server.com"
    tls_pubkey_pinset:
      - digest: "sha256"
        value: YOUR_CERT_PIN
```

## 安全建议

### 1. 路径 Token 选择

- ✅ 使用足够长和随机的字符串（至少 32 字符）
- ✅ 使用加密安全的随机数生成器
- ❌ 避免使用可预测的 token（如日期、用户名等）
- ❌ 避免在代码或配置文件中硬编码 token

**生成安全 token 示例：**
```bash
# Linux/Mac
openssl rand -hex 32

# 或使用 Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

### 2. 配置建议

- 将插件放在处理链的前端，尽早过滤
- 对于公网服务，优先使用路径白名单
- 定期轮换路径 token
- 监控和记录被拒绝的请求

### 3. 日志监控

插件会记录以下信息：
- 白名单加载的 IP 和路径数量
- 匹配失败时的警告日志

建议配置日志监控，及时发现异常访问。

## 故障排查

### 问题 1：所有请求都被拒绝

**可能原因：**
- 白名单配置错误
- 客户端 IP 无法获取（检查反向代理配置）
- 路径不匹配（检查路径是否正确）

**解决方法：**
```yaml
# 临时添加调试：允许所有内网 IP
whitelist:
  - "0.0.0.0/0"  # 仅用于调试，生产环境不要使用
```

### 问题 2：路径匹配失败

**检查项：**
- 路径是否包含前导斜杠
- 路径是否完全匹配（区分大小写）
- 检查实际请求路径（查看日志）

### 问题 3：IP 白名单不生效

**可能原因：**
- 客户端 IP 被反向代理隐藏
- 需要配置 `SrcIPHeader` 选项

**解决方法：**
在服务器配置中设置正确的 IP 头：
```yaml
servers:
  - exec: main_sequence
    listeners:
      - protocol: https
        addr: ":443"
        cert: cert.pem
        key: key.pem
    http_opts:
      src_ip_header: "X-Real-IP"  # 或 "X-Forwarded-For"
```

## 性能考虑

- IP 匹配使用高效的二分查找算法
- 路径匹配使用哈希表，O(1) 时间复杂度
- 插件开销极小，适合高并发场景
- 建议将插件放在处理链前端，减少后续处理开销

## ECS 功能说明

### 什么是 ECS？

ECS (EDNS Client Subnet) 是 DNS 的一个扩展，允许 DNS 查询携带客户端的子网信息，使上游 DNS 服务器能够返回更准确的地理位置相关的 DNS 响应。

### 路径 ECS 功能

插件支持为不同的路径自动添加不同的 ECS IP 地址：

- **自动添加**：当请求路径匹配 `path_ecs` 配置时，自动添加对应的 ECS 信息
- **智能选择**：根据查询类型（A/AAAA）自动选择 IPv4 或 IPv6
- **不覆盖**：如果查询中已有 ECS，不会覆盖（保护客户端隐私）

### ECS IP 选择逻辑

1. **A 查询（IPv4）**：优先使用配置的 IPv4，如果没有则使用 IPv6
2. **AAAA 查询（IPv6）**：优先使用配置的 IPv6，如果没有则使用 IPv4
3. **其他查询类型**：优先使用 IPv4，如果没有则使用 IPv6

### 运营商 IP 获取

可以使用以下方法获取各运营商的 IP 地址：

1. **使用真实运营商 IP**：从运营商网络获取真实 IP 地址
2. **使用测试 IP**：使用运营商提供的测试 IP 段
3. **使用 CDN IP**：使用对应运营商的 CDN 节点 IP

**示例运营商 IP（仅供参考，请使用实际 IP）：**
- 电信：`1.2.3.4` / `2001:db8::1`
- 移动：`5.6.7.8` / `2001:db8::2`
- 联通：`9.10.11.12` / `2001:db8::3`
- 广电：`13.14.15.16` / `2001:db8::4`

## 更新日志

- **v1.1.0**: 新增路径 ECS 功能，支持为不同路径自动添加不同的 ECS IP
- **v1.0.0**: 初始版本，支持 IP 和路径白名单
- 支持 OR 和 AND 匹配模式
- 支持从数据提供者动态加载

## 相关资源

- [mosdns-x Wiki](https://github.com/pmkol/mosdns-x/wiki)
- [DoH 协议规范](https://www.rfc-editor.org/rfc/rfc8484.html)
- [GitHub Issues](https://github.com/pmkol/mosdns-x/issues)

