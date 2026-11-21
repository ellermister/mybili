package version

// Version 是程序版本，不从环境变量读取。
// 打 tag 构建时，Dockerfile 会把这一行替换成 tag 提供的版本。
const Version = "2.0.0"
