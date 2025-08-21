## 🎥 Mybili

**bilibili 收藏夹下载工具**

主要是解决翻看收藏夹时，很多视频莫名其妙不见了的现象，也不知道原来的视频标题和内容，进而无法溯源和寻找备份。

**该工具能够将你的收藏夹全部备份下来。**

### 🛠️ 功能特点

- ⏰ 定时5分钟获取你的收藏夹所有视频，缓存标题、描述、封面等重要信息
- 🚀 自动通过队列，将你收藏的视频按照最高画质下载一份到本地
- 📺 提供友好的 web 页面展示你的收藏夹列表信息，以及进行在线播放预览
- 🎯 支持视频分P和弹幕全量缓存到本地，可以在线播放以及装载弹幕预览
- 📂 支持生成多媒体可读目录文件，适配 emby 等媒体库

## 🖼️ 预览

![preview](./preview.png)

## 📚 使用方法

该演示以最公共简单的方式创建一个服务，让你能够快速的体验到，你可以根据实际的需求和现实修改其中配置和部署方式。

- 💾 程序依赖 redis 来进行必要的缓存和分发处理异步任务
- 💿 默认将采用 sqlite 数据库存储收藏夹、视频、弹幕等文本型信息
- 📂 视频、图片将按照规则存储到你预设的本地文件系统中

### 🐳 1. 通过 docker 部署于你 nas

#### 📁 创建存储目录
建议为你将要缓存的视频资源创建目录，稍后将为其映射路径而不是通过 docker-compose 管理，目的是更方便的允许外部使用。

| 💡 如果你熟悉 Docker 配置，可以自由更改

```bash
mkdir /mnt/user/mybili/data -p
```

#### 📝 创建 docker 服务配置

创建文件 `/mnt/user/mybili/docker-compose.yml`

- 🔌 容器内部端口 80(http) 和 443(https) 都可以使用
- ⚙️ 下列 env 环境变量均可以参考 .env.example 文件内有的项进行覆盖，但默认只建议修改示例中内容

```yml
services: 
    mybili:
        image: ellermister/mybili
        ports:
            - "5151:80"
        volumes:
            - "./data:/app/storage/app/public"
            - db-data:/data

        environment:
            REDIS_HOST: redis
            REDIS_DB: 3
            DB_CONNECTION: sqlite
            DB_DATABASE: /data/database.sqlite
    redis:
        image: redis
        volumes:
            - redis-data:/data
        command: redis-server --save 60 1 --loglevel warning

volumes:
  db-data:
  redis-data:
```

#### 🚀 一键启动
```bash
docker-compose up -d
```

#### 🛑 停止运行 (如果你需要)
```bash
docker-compose down
```

### 🍪 2. 获取 cookie

你有两个方案可选其一

#### 📌 方案1 - 手动

1. 🔌 在你的浏览器安装插件
   [Get cookies.txt LOCALLY](https://chrome.google.com/webstore/detail/cclelndahbckbenkjhflpdbgdldlbecc)

2. 📤 在你登录哔哩哔哩后，通过插件导出 cookie 文件。需要格式为：`Netscape`

3. 🌐 访问 `http://your-ip:5151/cookie`

4. ⬆️ 上传 cookie 文件，稍后将自动开始同步你的收藏夹了！🍡🍡🍡

#### 🤖 方案2 - 自动

由于方案1上传 cookie 会话之后，会在几天之后自动过期, 无法实现长期的自动同步。

原因是因为登录的网页版 bilibili 在同期使用时，重新获取了新的短期 token，而 mybili 并没有更新，也没有机制去自动获取新的 token，如果 mybili 自己去获取新的 token 也会导致你的网页版本掉线。无论是从实现复杂度还是使用体验来论都不好。

目前参考上述插件获取 cookie 内容，加以加工，制作了一个简单的自动同步 cookie chrome 扩展，只需要填写你的 mybili 网页地址。就能够实现自动无感知自动同步 cookie 到 mybili。

https://github.com/ellermister/mybili-cookie

1. 📥 打开项目地址，点击 "Code" -> "Download ZIP" 下载项目
2. 📦 将下载的项目解压到本地目录，长期使用请合理安排目录位置，如 `C:\mybili-cookie`
3. 🔧 打开 chrome 浏览器 `chrome://extensions/` 打开 "开发人员模式"
4. 📂 加载解压缩的扩展 选择目录 `C:\mybili-cookie` 以开启扩展
5. ⚙️ 点击新安装的扩展，在弹出的 popup.html 页面里填写你的 mybili 地址，截至到端口即可, 如 `http://192.168.1.200:5151`


### ⚙️ 3. 配置同步

在网页设置页面配置下载选项（默认关闭同步）。

访问：`http://your-ip:5151/settings`

**主要设置项：**
- **收藏夹同步**：总开关，每5分钟同步收藏夹信息
- **多P下载**：下载视频所有分段，否则仅下载第一个
- **弹幕下载**：下载视频弹幕到数据库
- **视频下载**：允许触发视频文件下载


### 📝 4. 日志排查

在容器内部，存储了多份日志，来源于不同的服务产生的文件。
```bash
/app # ls /var/log
queue.log.0        schedule.log.0     supervisord.log.0  web.log.0
```

#### 🌐 Web 服务日志
网页不通或者异常报错，可以查看 laravel 的日志
```bash
tail -f /app/storage/logs/laravel.log
```

## 🔄 更新说明

你可以在 github 或 docker hub 检视如果存在新版本，可以直接通过拉取最新镜像进行更新。

### ⚠️ 注意事项

1. 🔄 当前版本已经废除 redis 存储持久化数据，只用于写入临时缓存和异步队列用途
2. ⚙️ 当前版本指引中已经废除 .env 进行配置，采用环境变量进行覆盖配置，如果你有配置可以在 docker-compose.yml 中移除 .env 条目
3. 🍪 当前版本指引中已经废除 cookie.txt 文件映射，采用数据库进行存储，如果你有配置可以在 docker-compose.yml 中移除 cookie.txt 条目

## 💓 支持本项目

如果你喜欢这个项目，或者它对你有帮助，请考虑支持我！

### 💰 赞助方式

你可以通过以下方式支持这个项目：

- **Buy me a coffee**: [买一杯咖啡](https://buymeacoffee.com/ellermister)
- **USDT**: TRC20 `TRjWTbPfQBhHawCD8DrfLGa8ECbhPP6F3b`
- **LTC**: Litecoin `LdH6SxbAq3No9P4zaNR2aGgH9Kr9yfuGHi`

任何形式的支持都将帮助我继续改进和维护这个项目，非常感谢！

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=ellermister/mybili&type=Date)](https://www.star-history.com/#ellermister/mybili&Date)