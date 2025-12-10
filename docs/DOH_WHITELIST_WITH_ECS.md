# doh_whitelist 插件与 ECS 配置指南

## 概述

`doh_whitelist` 插件本身**不实现 ECS（EDNS Client Subnet）功能**，但可以**配合 ECS 插件一起使用**。

- **doh_whitelist**：用于限制 DoH 请求的访问（白名单鉴权）
- **ecs 插件**：用于在 DNS 查询中添加 ECS 信息，帮助上游 DNS 返回更准确的地理位置相关响应

## ECS 插件简介

ECS（EDNS Client Subnet）是 DNS 的一个扩展，允许 DNS 查询携带客户端的子网信息，使上游 DNS 服务器能够返回更准确的地理位置相关的 DNS 响应。

### ECS 插件功能

- 自动使用客户端 IP 添加 ECS 信息
- 支持预设 IP 地址
- 支持自定义子网掩码
- 自动从响应中移除 ECS 信息（保护隐私）

## 配置示例

### 示例 1：DoH 白名单 + ECS（推荐配置）

同时使用白名单限制访问和 ECS 优化响应：

```yaml
plugins:
  # DoH 白名单插件：限制访问
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      path_list:
        - "/dns-query/token123"
      # 或使用 IP 白名单
      # whitelist:
      #   - "192.168.1.0/24"

  # ECS 插件：添加客户端子网信息
  - tag: ecs
    type: ecs
    args:
      auto: true           # 自动使用客户端 IP
      mask4: 24            # IPv4 子网掩码（默认 24）
      mask6: 48            # IPv6 子网掩码（默认 48）
      force_overwrite: false  # 不覆盖已有的 ECS

  # 主处理序列
  - tag: main_sequence
    type: sequence
    args:
      exec:
        - doh_whitelist    # 1. 先检查白名单
        - ecs              # 2. 添加 ECS 信息
        - cache            # 3. 缓存
        - forward          # 4. 转发到上游
```

**工作流程：**
1. DoH 请求到达
2. `doh_whitelist` 检查路径或 IP 是否在白名单中
3. 如果通过，`ecs` 插件添加客户端子网信息到查询
4. 继续后续处理（缓存、转发等）

### 示例 2：仅内网使用 ECS

内网用户使用 ECS，公网用户不使用：

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      whitelist:
        - "192.168.1.0/24"      # 内网 IP
        - "10.0.0.0/8"          # 更大的内网段
      path_list:
        - "/dns-query/public"   # 公网路径（不使用 ECS）
      require_both: false

  - tag: ecs_internal
    type: ecs
    args:
      auto: true
      mask4: 24
      mask6: 48

  - tag: main_sequence
    type: sequence
    args:
      exec:
        - doh_whitelist
        - if:
            # 如果是内网 IP，添加 ECS
            - exec: ecs_internal
            # 否则跳过
        - cache
        - forward
```

### 示例 3：使用预设 IP 的 ECS

不使用客户端真实 IP，而是使用预设的 IP 地址：

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      path_list:
        - "/dns-query/token123"

  - tag: ecs_preset
    type: ecs
    args:
      auto: false           # 不使用客户端 IP
      ipv4: "1.2.3.4"       # 预设 IPv4
      ipv6: "2001:db8::1"   # 预设 IPv6
      mask4: 24
      mask6: 48

  - tag: main_sequence
    type: sequence
    args:
      exec:
        - doh_whitelist
        - ecs_preset
        - cache
        - forward
```

### 示例 4：条件使用 ECS

根据查询类型决定是否使用 ECS：

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      path_list:
        - "/dns-query/token123"

  - tag: ecs_auto
    type: ecs
    args:
      auto: true
      mask4: 24
      mask6: 48

  - tag: main_sequence
    type: sequence
    args:
      exec:
        - doh_whitelist
        - if:
            # 仅对 A 和 AAAA 查询使用 ECS
            - if:
                - qtype: [A, AAAA]
                - exec: ecs_auto
        - cache
        - forward
```

## ECS 插件参数说明

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `auto` | `bool` | 否 | 是否自动使用客户端 IP（默认：false） |
| `ipv4` | `string` | 否 | 预设的 IPv4 地址（auto=false 时使用） |
| `ipv6` | `string` | 否 | 预设的 IPv6 地址（auto=false 时使用） |
| `mask4` | `int` | 否 | IPv4 子网掩码（默认：24） |
| `mask6` | `int` | 否 | IPv6 子网掩码（默认：48） |
| `force_overwrite` | `bool` | 否 | 是否强制覆盖已有的 ECS（默认：false） |

## 使用场景

### 场景 1：CDN 优化

使用 ECS 帮助 CDN DNS 返回更近的节点：

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      path_list:
        - "/dns-query/cdn-token"

  - tag: ecs
    type: ecs
    args:
      auto: true
      mask4: 24    # 使用 /24 掩码保护隐私
      mask6: 48    # 使用 /48 掩码保护隐私

  - tag: forward_cdn
    type: forward
    args:
      upstream:
        - addr: "https://cloudflare-dns.com/dns-query"
```

### 场景 2：地理位置相关服务

某些 DNS 服务需要客户端位置信息：

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      whitelist:
        - "192.168.1.0/24"

  - tag: ecs
    type: ecs
    args:
      auto: true
      mask4: 24
      mask6: 48
```

### 场景 3：隐私保护

使用较大的子网掩码保护客户端隐私：

```yaml
plugins:
  - tag: ecs
    type: ecs
    args:
      auto: true
      mask4: 16    # 使用 /16 掩码，只暴露网段
      mask6: 32    # 使用 /32 掩码，只暴露网段
```

## 注意事项

### 1. 插件顺序很重要

ECS 插件应该放在白名单检查**之后**，这样只有通过白名单的请求才会添加 ECS：

```yaml
exec:
  - doh_whitelist    # ✅ 先检查白名单
  - ecs              # ✅ 然后添加 ECS
  - forward          # ✅ 最后转发
```

### 2. 隐私考虑

- ECS 会向上游 DNS 服务器暴露客户端子网信息
- 建议使用较大的子网掩码（如 /24 或 /16）保护隐私
- 某些场景下可能不需要 ECS，可以禁用

### 3. 性能影响

- ECS 插件开销很小，几乎不影响性能
- 白名单检查在 ECS 之前，可以提前过滤请求

### 4. 兼容性

- 不是所有上游 DNS 服务器都支持 ECS
- 如果上游不支持，ECS 信息会被忽略
- 某些 DNS 服务器可能会移除 ECS 信息

## 测试 ECS 是否生效

### 方法 1：使用 dig 查看

```bash
# 查看查询中的 ECS 信息
dig @your-server.com example.com +subnet=1.2.3.4/24
```

### 方法 2：查看日志

启用调试日志，查看 ECS 是否被添加：

```yaml
log:
  level: debug
```

### 方法 3：使用 tcpdump 抓包

```bash
# 抓取 DNS 查询包
tcpdump -i any -n port 53 -v
```

查看 EDNS0 选项中是否包含 SUBNET 信息。

## 常见问题

### Q1: doh_whitelist 插件本身能实现 ECS 吗？

**A:** 不能。`doh_whitelist` 插件只负责白名单检查，不涉及 ECS 功能。需要配合 `ecs` 插件使用。

### Q2: 为什么需要同时使用两个插件？

**A:** 
- `doh_whitelist`：控制**谁可以访问**（安全）
- `ecs`：优化**DNS 响应质量**（性能）

两者功能不同，可以互补使用。

### Q3: ECS 会影响白名单检查吗？

**A:** 不会。白名单检查基于客户端 IP 和路径，ECS 只是添加到查询中的额外信息，不影响白名单逻辑。

### Q4: 如何禁用 ECS？

**A:** 不使用 `ecs` 插件，或者使用 `_no_ecs` 预设插件移除 ECS：

```yaml
- tag: remove_ecs
  type: sequence
  args:
    exec:
      - _no_ecs  # 移除 ECS
```

## 总结

- ✅ `doh_whitelist` 插件本身**不实现 ECS**
- ✅ 可以**配合 `ecs` 插件**一起使用
- ✅ 建议将白名单检查放在 ECS 之前
- ✅ 根据实际需求决定是否使用 ECS
- ✅ 注意隐私保护，使用合适的子网掩码

