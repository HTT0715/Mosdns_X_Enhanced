# Git 推送代码到 GitHub 使用指南

## 前置准备

### 1. 安装 Git

**Windows:**
- 下载地址：https://git-scm.com/download/win
- 或使用包管理器：`winget install Git.Git`

**验证安装：**
```bash
git --version
```

### 2. 配置 Git 用户信息

```bash
git config --global user.name "你的名字"
git config --global user.email "your.email@example.com"
```

### 3. 配置 GitHub 认证

**方法 1：使用 Personal Access Token (推荐)**
1. 登录 GitHub
2. 进入 Settings → Developer settings → Personal access tokens → Tokens (classic)
3. 生成新 token，勾选 `repo` 权限
4. 保存 token（只显示一次）

**方法 2：使用 SSH 密钥**
```bash
# 生成 SSH 密钥
ssh-keygen -t ed25519 -C "your.email@example.com"

# 复制公钥内容
cat ~/.ssh/id_ed25519.pub

# 添加到 GitHub: Settings → SSH and GPG keys → New SSH key
```

## 推送代码到指定仓库和分支

### 步骤 1：检查当前状态

```bash
# 查看当前分支
git branch

# 查看远程仓库配置
git remote -v

# 查看文件变更状态
git status
```

### 步骤 2：添加远程仓库（如果还没有）

**如果这是新仓库：**
```bash
# 添加远程仓库
git remote add origin https://github.com/你的用户名/仓库名.git

# 或使用 SSH
git remote add origin git@github.com:你的用户名/仓库名.git
```

**如果已有远程仓库，但想更换：**
```bash
# 查看现有远程仓库
git remote -v

# 删除现有远程仓库
git remote remove origin

# 添加新的远程仓库
git remote add origin https://github.com/你的用户名/仓库名.git
```

### 步骤 3：提交代码

```bash
# 查看所有变更
git status

# 添加所有变更的文件
git add .

# 或添加特定文件
git add plugin/executable/doh_whitelist/doh_whitelist.go
git add README.md
git add .github/workflows/build.yml

# 提交变更（使用有意义的提交信息）
git commit -m "Add doh_whitelist plugin with IP and path whitelist support"
```

### 步骤 4：推送到指定分支

**推送到 main 分支：**
```bash
# 如果当前在 main 分支
git push origin main

# 如果当前在其他分支，想推送到 main
git push origin 当前分支名:main
```

**推送到 master 分支：**
```bash
git push origin master
```

**推送到自定义分支：**
```bash
# 创建并切换到新分支
git checkout -b feature/doh-whitelist

# 或切换到已存在的分支
git checkout 分支名

# 推送当前分支
git push origin feature/doh-whitelist

# 首次推送新分支需要设置上游
git push -u origin feature/doh-whitelist
```

**强制推送（谨慎使用）：**
```bash
# 仅在确定要覆盖远程分支时使用
git push -f origin 分支名
```

## 完整示例流程

```bash
# 1. 进入项目目录
cd C:\Users\yubo1\Desktop\mosdns-x-main

# 2. 检查状态
git status

# 3. 添加所有变更
git add .

# 4. 提交变更
git commit -m "Add doh_whitelist plugin and GitHub Actions build workflow"

# 5. 如果还没有设置远程仓库
git remote add origin https://github.com/你的用户名/mosdns-x.git

# 6. 推送到 main 分支
git push origin main

# 如果提示需要认证，输入用户名和 Personal Access Token
```

## 在哪里查看

### 1. 查看本地 Git 状态

```bash
# 查看当前分支
git branch

# 查看所有分支（包括远程）
git branch -a

# 查看提交历史
git log --oneline

# 查看远程仓库配置
git remote -v

# 查看文件变更
git status
git diff
```

### 2. 在 GitHub 上查看

**仓库页面：**
- 访问：`https://github.com/你的用户名/仓库名`
- 查看代码、提交历史、分支等

**Actions 页面：**
- 访问：`https://github.com/你的用户名/仓库名/actions`
- 查看构建状态、下载构建产物

**Releases 页面：**
- 访问：`https://github.com/你的用户名/仓库名/releases`
- 查看发布的版本和下载链接

**分支页面：**
- 访问：`https://github.com/你的用户名/仓库名/branches`
- 查看所有分支

### 3. 在 VS Code 中查看

如果使用 VS Code：
- 左侧边栏的 "源代码管理" 图标（Ctrl+Shift+G）
- 可以查看变更、提交、推送等

### 4. 在命令行查看

```bash
# 查看远程分支
git branch -r

# 查看所有分支的跟踪关系
git branch -vv

# 查看远程仓库信息
git remote show origin
```

## 常见问题

### 问题 1：认证失败

**错误信息：** `fatal: Authentication failed`

**解决方法：**
- 使用 Personal Access Token 代替密码
- 或配置 SSH 密钥

### 问题 2：分支不存在

**错误信息：** `error: src refspec main does not match any`

**解决方法：**
```bash
# 先创建并提交一些内容
git add .
git commit -m "Initial commit"
git push -u origin main
```

### 问题 3：远程分支已更新

**错误信息：** `Updates were rejected because the remote contains work`

**解决方法：**
```bash
# 先拉取远程更新
git pull origin main

# 解决冲突后再次推送
git push origin main
```

### 问题 4：查看推送历史

```bash
# 查看推送日志
git log origin/main..HEAD

# 查看未推送的提交
git log @{u}..
```

## 快速参考命令

```bash
# 初始化仓库（如果是新项目）
git init

# 添加远程仓库
git remote add origin <仓库URL>

# 查看远程仓库
git remote -v

# 添加文件
git add .

# 提交
git commit -m "提交信息"

# 推送到远程
git push origin 分支名

# 拉取远程更新
git pull origin 分支名

# 创建新分支
git checkout -b 新分支名

# 切换分支
git checkout 分支名

# 查看状态
git status

# 查看提交历史
git log --oneline --graph
```

## 推送后验证

推送成功后：

1. **在 GitHub 仓库页面查看：**
   - 刷新仓库页面，应该能看到新的提交
   - 查看 "commits" 标签页确认提交已推送

2. **在 Actions 页面查看：**
   - 进入 Actions 标签页
   - 应该能看到自动触发的构建工作流
   - 等待构建完成后可以下载构建产物

3. **检查分支：**
   - 在仓库页面点击分支下拉菜单
   - 确认你的分支已存在

