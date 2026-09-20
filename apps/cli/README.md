# Beep CLI

`beep` 是 Beep 的命令行工具与后台守护进程服务。

## 本地开发 (`beep-local`)

`beep-local` 是本地开发使用的，通过 mise 配置加载，直接运行当前 Go 代码：

```bash
beep-local --help
```

## Install current tree locally (`mise cli:install`)

Build the current source as the real `beep` binary and install it into your existing install directory (usually `~/.local/bin`) so you can dogfood before cutting a release. This does not use GitHub Releases.

```bash
mise cli:install
beep version
beep service restart
```

Set `INSTALL_DIR` to override the destination. In this repo, mise puts `./bin` earlier on `PATH` than `~/.local/bin`, so `which beep` may still point at the repo build; `cli:install` is meant to overwrite the user-installed binary.
