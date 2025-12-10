# 路径 ECS 配置示例 - 运营商分流

## 概述

通过配置 `path_ecs`，可以为不同的路径自动添加不同运营商的 ECS IP，实现运营商分流和 CDN 优化。

## 完整配置示例

### 示例：电信、移动、联通、广电四线分流

```yaml
plugins:
  - tag: doh_whitelist
    type: doh_whitelist
    args:
      # 路径 ECS 配置：不同路径使用不同运营商的 IP
      path_ecs:
        # 电信线路
        "/dns-query/telecom":
          ipv4: "1.2.3.4"           # 电信 IPv4
          ipv6: "2001:db8:1::1"     # 电信 IPv6（可选）
          mask4: 24                 # IPv4 掩码
          mask6: 48                 # IPv6 掩码
        
        # 移动线路
        "/dns-query/mobile":
          ipv4: "5.6.7.8"
          ipv6: "2001:db8:2::1"
          mask4: 24
          mask6: 48
        
        # 联通线路
        "/dns-query/unicom":
          ipv4: "9.10.11.12"
          ipv6: "2001:db8:3::1"
          mask4: 24
          mask6: 48
        
        # 广电路线
        "/dns-query/cable":
          ipv4: "13.14.15.16"
          mask4: 24

  - tag: main_sequence
    type: sequence
    args:
      exec:
        - doh_whitelist
        - cache
        - forward
```

## 使用方式

### 客户端配置

**电信用户：**
```
https://your-server.com/dns-query/telecom
```

**移动用户：**
```
https://your-server.com/dns-query/mobile
```

**联通用户：**
```
https://your-server.com/dns-query/unicom
```

**广电用户：**
```
https://your-server.com/dns-query/cable
```

### 工作原理

1. 客户端使用对应运营商的路径访问 DoH 服务
2. 插件自动识别路径，添加对应运营商的 ECS IP
3. 上游 DNS 服务器根据 ECS IP 返回对应运营商的最优节点
4. 客户端获得更快的 DNS 解析速度

## 获取运营商 IP

### 方法 1：使用真实运营商 IP

从运营商网络环境获取真实 IP：

```bash
# 在对应运营商网络环境下执行
curl ifconfig.me
# 或
curl ip.sb
```

### 方法 2：使用运营商测试 IP 段

参考各运营商提供的测试 IP 段（需要查询最新信息）。

### 方法 3：使用 CDN 节点 IP

使用对应运营商的 CDN 节点 IP 地址。

## 配置建议

### 掩码选择

- **IPv4 掩码**：建议使用 `/24`（保护隐私，同时提供足够的位置信息）
- **IPv6 掩码**：建议使用 `/48`（保护隐私，同时提供足够的位置信息）

### 隐私保护

- 使用较大的掩码（如 `/24` 或 `/16`）可以保护客户端隐私
- 只暴露网段信息，不暴露具体 IP 地址

### 性能优化

- 为每个运营商配置对应的 IPv4 和 IPv6
- 根据实际网络环境调整掩码大小

## 测试验证

### 使用 dig 测试

```bash
# 测试电信路径
dig @https://your-server.com/dns-query/telecom example.com

# 测试移动路径
dig @https://your-server.com/dns-query/mobile example.com
```

### 查看 ECS 信息

启用调试日志查看 ECS 是否被正确添加：

```yaml
log:
  level: debug
```

日志中会显示：
```
added ECS for path path=/dns-query/telecom ecs_ip=1.2.3.4 mask=24
```

## 常见问题

### Q: 如何知道 ECS 是否生效？

A: 启用调试日志，查看是否有 "added ECS for path" 的日志输出。

### Q: 可以同时配置路径白名单和路径 ECS 吗？

A: 可以。路径白名单控制访问权限，路径 ECS 控制 ECS 信息，两者可以同时使用。

### Q: 如果路径不匹配 path_ecs 配置会怎样？

A: 不会添加 ECS，查询会正常处理，只是没有 ECS 信息。

### Q: 如何获取各运营商的真实 IP？

A: 在对应运营商的网络环境下，使用 `curl ifconfig.me` 或类似工具获取。

