## Mosdns-x

Mosdns-x 是一个用 Go 编写的高性能 DNS 转发器，支持运行插件流水线，用户可以按需定制 DNS 处理逻辑。

**支持监听与请求以下类型的 DNS：**

* UDP
* TCP
* DNS over TLS - DoT
* DNS over QUIC - DoQ
* DNS over HTTP/2 - DoH
* DNS over HTTP/3 - DoH3

功能概述、配置方式、教程，详见：[wiki](https://github.com/pmkol/mosdns-x/wiki)

下载预编译文件、更新日志，详见：[release](https://github.com/pmkol/mosdns-x/releases)

## 插件使用示例

### doh_whitelist - DoH 白名单客户端插件

`doh_whitelist` 插件用于限制 DoH (DNS over HTTPS) 请求，支持基于 IP 地址和 URL 路径的双重白名单鉴权。

**功能特性：**
- 仅对 DoH 请求生效（自动识别 HTTPS/H2/H3 协议）
- 支持 IP 地址白名单（单个 IP 或 CIDR 网段）
- 支持 URL 路径白名单（通过路径 token 鉴权）
- 支持 IP 和路径的灵活组合（OR 或 AND 逻辑）
- 支持从数据提供者动态加载 IP 白名单
- 非 DoH 请求自动通过，不影响其他协议
- 可配置拒绝响应码

**配置示例：**

**示例 1：仅使用 IP 白名单**
```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      whitelist:
        - "192.168.1.0/24"      # 允许内网访问
        - "10.0.0.1"            # 允许特定 IP
        - "2001:db8::/32"       # 允许 IPv6 网段
        # - "provider:my_provider:trusted_ips"  # 从数据提供者加载
```

**示例 2：仅使用路径白名单（推荐用于公网服务）**
```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      # 路径白名单：允许访问 /dns-query/token123 等路径
      path_list:
        - "/dns-query/token123"
        - "/dns-query/secret-key-456"
        - "/dns-query/my-custom-path"
      # 客户端访问示例：
      # https://your-server.com/dns-query/token123
```

**示例 3：IP 或路径任一匹配即可（默认模式）**
```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      whitelist:
        - "192.168.1.0/24"      # 内网 IP 白名单
      path_list:
        - "/dns-query/public-token"  # 公网路径白名单
      # require_both: false  # 默认值，IP 或路径任一匹配即可
      # 这样配置后：
      # - 内网用户（192.168.1.0/24）可以直接访问任何路径
      # - 公网用户必须使用正确的路径 token 才能访问
```

**示例 4：IP 和路径必须同时匹配（严格模式）**
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
      # 这样配置后，只有内网 IP 且使用正确路径才能访问
```

**完整配置参数：**
```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      # IP 白名单（可选）
      whitelist:
        - "192.168.1.0/24"      # CIDR 网段
        - "10.0.0.1"            # 单个 IP
        - "2001:db8::/32"       # IPv6 CIDR
        # - "provider:my_provider:trusted_ips"  # 从数据提供者加载
      
      # 路径白名单（可选）
      path_list:
        - "/dns-query/token123"     # 完整路径
        - "/dns-query/secret-key"   # 支持多个路径
        - "/custom-path"            # 自动添加前导斜杠
      
      # 匹配模式（可选，默认 false）
      # false: IP 或路径任一匹配即可（OR 逻辑）
      # true:  IP 和路径必须同时匹配（AND 逻辑）
      require_both: false
      
      # 拒绝响应码（可选，默认 5 = REFUSED）
      rcode: 5
      # 其他常用响应码：
      #   0 = NOERROR
      #   2 = SERVFAIL
      #   3 = NXDOMAIN
      #   5 = REFUSED (默认)
```

**使用场景：**
- **仅 IP 白名单**：限制 DoH 服务仅对特定客户端 IP 开放
- **仅路径白名单**：公网服务，通过 URL 路径 token 实现鉴权（如：`https://dns.example.com/dns-query/your-secret-token`）
- **IP + 路径组合**：内网用户直接访问，公网用户需要 token；或要求同时满足 IP 和路径条件
- 防止未授权访问 DoH 服务
- 配合反向代理使用，实现灵活的访问控制

**注意事项：**
- 插件仅检查 DoH 请求，UDP/TCP/DoT/DoQ 等其他协议不受影响
- 至少需要配置 `whitelist` 或 `path_list` 其中之一
- 如果只配置了 IP 白名单且客户端 IP 无法获取（无效地址），请求将被拒绝
- 路径匹配是精确匹配（自动规范化处理），不支持通配符
- 建议将此插件放在处理链的前端，以尽早过滤未授权请求
- 路径 token 建议使用足够复杂和随机的字符串，以提高安全性

**📖 详细使用文档：** 查看 [doh_whitelist 插件完整文档](https://github.com/HTT0715/Mosdns_X_Enhanced/blob/main/docs/DOE_WHITELIST_WITH_ECS.md) 获取更多配置示例、客户端配置方法、故障排查和安全建议。

#### 电报社区：

**[Mosdns-x Group](https://t.me/mosdns)**

#### 关联项目：

**[easymosdns](https://github.com/pmkol/easymosdns)**

适用于 Linux 的辅助脚本。借助 Mosdns-x，仅需几分钟即可搭建一台支持 ECS 的无污染 DNS 服务器。内置中国大陆地区的优化规则，满足DNS日常使用场景，开箱即用。

**[mosdns-v4](https://github.com/IrineSistiana/mosdns/tree/v4)**

一个插件化的 DNS 转发器。是 Mosdns-x 的上游项目。
