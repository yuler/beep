# Beep CLI

`beep` 是 Beep 的命令行工具与后台守护进程服务。

## 本地开发 (`beep-local`)

`beep-local` 是本地开发使用的，通过 mise 配置加载，直接运行当前 Go 代码：

```bash
beep-local --help
```

## 本地安装当前 tree（`mise cli:install`）

把当前源码编成正式名字的 `beep`，装到已有安装目录（通常是 `~/.local/bin`），方便自己先用一段时间再发版。不走 GitHub Release。

```bash
mise cli:install
beep version
beep service restart
```

`INSTALL_DIR` 可覆盖目标目录。仓库里 `mise` 的 `./bin` 排在 `~/.local/bin` 后面，所以 `which beep` 仍可能指向用户安装；`cli:install` 就是为了覆盖那一份。
