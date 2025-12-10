# GitHub Actions 使用说明

本项目使用 GitHub Actions 进行自动化构建和发布。

## 工作流说明

### 1. build.yml - 自动构建工作流

**触发条件：**
- Push 到 `main` 或 `master` 分支
- 创建以 `v` 开头的 tag（如 `v1.0.0`）
- 创建 Pull Request
- 手动触发（workflow_dispatch）

**功能：**
- 自动构建多个平台的二进制文件
- 上传构建产物到 GitHub Actions Artifacts
- 当创建 tag 时，自动创建 GitHub Release 并上传所有平台的构建文件

**支持的平台：**
- Linux (amd64, arm64, arm)
- Windows (amd64)
- macOS (amd64, arm64)

### 2. release.yml - 完整发布工作流

**触发条件：**
- 手动触发（workflow_dispatch）

**功能：**
- 构建所有支持的平台（包括更多架构如 mips、ppc64le 等）
- 创建预发布版本（prerelease）

### 3. test.yml - 测试工作流

**触发条件：**
- Push 代码
- 创建 Pull Request

**功能：**
- 测试代码是否能正常编译

## 使用方法

### 自动构建

1. **Push 代码到 main/master 分支**
   - 工作流会自动触发
   - 构建产物会上传到 Actions Artifacts
   - 可以在 Actions 页面下载构建的二进制文件

2. **创建 Release Tag**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
   - 会自动创建 GitHub Release
   - 上传所有平台的构建文件
   - 生成 Release 说明

### 手动触发构建

1. 进入 GitHub 仓库页面
2. 点击 "Actions" 标签
3. 选择 "Build mosdns-x" 工作流
4. 点击 "Run workflow"
5. 选择分支并运行

### 下载构建产物

1. 进入 GitHub 仓库页面
2. 点击 "Actions" 标签
3. 选择最新的工作流运行
4. 在 "Artifacts" 部分下载对应平台的二进制文件

### 发布 Release

1. 创建并推送 tag：
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. GitHub Actions 会自动：
   - 构建所有平台的二进制文件
   - 创建 GitHub Release
   - 上传构建文件到 Release

3. 可以在 Releases 页面查看和下载

## 构建配置

构建使用以下参数：
- `CGO_ENABLED=0` - 禁用 CGO，生成静态链接的二进制
- `GOEXPERIMENT=greenteagc` - 启用绿色 GC 实验特性
- 使用 `-ldflags` 注入版本和构建时间信息
- 使用 `-trimpath` 移除构建路径信息
- 使用 `-s -w` 减小二进制文件大小

## 注意事项

1. 确保 `release/` 目录在 `.gitignore` 中（已配置）
2. 构建产物会自动上传到 Artifacts，保留 30-90 天
3. Release 构建会包含所有支持的平台
4. 版本号格式：tag 使用 tag 名称，其他使用 `dev-日期-提交哈希`



