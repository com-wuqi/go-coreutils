# Go Coreutils 练习场

用 Go 复刻常见终端工具，练习命令行参数解析、流式 I/O、字节与 Unicode 处理、错误信息与退出码。

**这个仓库只提供脚手架：目录结构、行为规范、190 个验收测试。所有命令实现和共享解析器的 `Parse` 都留给你。**

## 当前状态

- `internal/<tool>/<tool>.go` 中只有一个 `Run` 存根，固定返回 `not implemented`。
- `internal/opt/opt.go` 中是共享解析器的 `Parse` 存根，同样待实现。
- `go test ./...` 当前故意全红（6 个工具包 + `internal/opt`）。
- 你按工具实现 `Run`（并先实现 `Parse`），对应包的测试就会逐项变绿。

## 环境要求

- Go 1.27 或更新版本（`go.mod` 声明 `go 1.27.1`）。
- 可选：PowerShell 7（用于 `scripts/check.ps1`）。
- 不需要网络，无第三方依赖。

## 快速开始

```powershell
go version
go build ./...
go test ./...                 # 初始：6 个工具包 + opt 全部失败
go test ./internal/echo -v    # 建议先从一个工具开始
go test ./internal/opt -v     # cat/head/tail/wc/tee 的共享解析器
go run ./cmd/echo hello       # 未实现时输出 "echo: not implemented"
```

初始失败的样子：

```text
--- FAIL: TestRun (0.00s)
    --- FAIL: TestRun/no_args_prints_newline (0.00s)
        testutil.go:56: exit code mismatch:
             got: 1
            want: 0
```

这不是脚手架坏了，而是实现尚未开始。

## 目录结构

```text
go-coreutils/
├── cmd/<tool>/main.go          # 只负责把 os 参数和标准流接到 Run
├── internal/<tool>/
│   ├── <tool>.go               # 你唯一需要实现的文件
│   ├── <tool>_test.go          # 验收测试，不要为了让测试通过而修改
│   └── testdata/               # 固定输入文件
├── internal/cli/               # 退出码、RunFunc、NotImplemented
├── internal/opt/               # 共享短选项扫描（Parse 存根 + 25 个测试）
├── internal/testutil/          # 测试辅助：捕获 I/O、断言、临时文件
├── docs/
│   ├── architecture.md         # 分层与统一契约
│   ├── testing.md              # 测试运行与编写指南
│   ├── progress.md             # 实现进度清单
│   └── specs/                  # 各工具与 opt 的行为规范（唯一准绳）
└── scripts/check.ps1           # gofmt + vet + build + test 一键检查
```

## 推荐学习顺序

| 顺序 | 目标                               | 主要练习                                   |
|------|------------------------------------|--------------------------------------------|
| 1    | [echo](docs/specs/echo.md)         | 参数拼接、短选项状态、转义与字节输出       |
| 2    | [opt（共享包）](docs/specs/opt.md) | 短选项扫描、附加值、`--`、错误契约         |
| 3    | [cat](docs/specs/cat.md)           | 流式读取、多输入、行号、行状态、错误后继续 |
| 4    | [tee](docs/specs/tee.md)           | 一次读取、多路写出、文件打开失败后的降级   |
| 5    | [head](docs/specs/head.md)         | 按行/按字节截断、文件头、选项覆盖顺序      |
| 6    | [tail](docs/specs/tail.md)         | 从尾部选择、`+N`、缓冲策略、末尾无换行     |
| 7    | [wc](docs/specs/wc.md)             | 计数、Unicode、固定列顺序、多文件汇总      |

## 每个工具的练习闭环

1. 阅读 `docs/specs/<tool>.md`，先只看目标、选项、行为三节。
2. 在 `internal/<tool>/<tool>.go` 中实现 `Run`；可以边实现边跑测试。
3. 运行 `go test ./internal/<tool> -v`，从第一个失败用例开始修。
4. 全部通过后，对照 spec 的“边界情况”手动补测几个输入。
5. 在 `docs/progress.md` 中勾掉对应项。
6. 想继续挑战时，完成 spec 末尾的“进阶练习”，并为它补充你自己的测试。

`internal/opt` 也使用同一闭环：读 [opt.md](docs/specs/opt.md) → 实现 `Parse` → `go test ./internal/opt -v`
。cat/head/tail/wc/tee 推荐复用它的解析结果。

规则：

- **测试是契约的可执行形式。** 不要修改 `_test.go` 或删减断言来迁就实现。
- 如果测试与 spec 冲突，以 spec 为准，同时修正测试与文档。
- 允许重构；测试只看 `Run` 的公开行为。

## 实现契约

每个工具都必须实现（`internal/cli` 中已定义 `RunFunc`）：

```go
package cli
import (
	"io"
)
type RunFunc func(args []string, stdin io.Reader, stdout, stderr io.Writer) int
```

- `args` 不包含程序名。
- 返回进程退出码；不要在 `internal` 包中调用 `os.Exit`、`log.Fatal` 或 `panic`。
- 不要关闭 `stdin`/`stdout`/`stderr`；自己打开的文件要关闭。
- 不要依赖当前工作目录之外的全局状态。

退出码：

| 退出码 | 含义                                         |
|--------|----------------------------------------------|
| 0      | 全部成功                                     |
| 1      | 运行时错误或部分失败（继续处理能处理的输入） |
| 2      | 参数/用法错误（解析失败，不处理任何操作数）  |

错误信息：

- 用法错误：`<prog>: unrecognized option: -z`、`<prog>: option requires an argument: -n`、`<prog>: invalid number: abc`。
- 文件错误：`<prog>: <路径>: <底层错误>`。
- 全部写到 stderr，以换行结尾。

参数解析（cat/head/tail/wc/tee）：

- 短选项可组合（`-nE`），带值选项可写 `-n5` 或 `-n 5`。
- `--` 之后全部是操作数；单独的 `-` 是操作数。
- 选项可以出现在操作数之后（GNU 风格混排）。
- 未知选项、缺少值、非法数字 → 退出码 2。
- 共享扫描器位于 `internal/opt`（类型、测试、`Parse` 存根已就绪）；工具可以复用它，也可以自行解析，验收测试只看行为。
- echo 是特例：只识别前置的 `-n`/`-e`/`-E`，其余按字面量处理。

完整约定见 [docs/architecture.md](docs/architecture.md)。

## 测试与检查

```powershell
go test ./...                       # 全部测试
go test ./internal/cat -v           # 单个工具，显示每个子测试
go test ./internal/opt -v           # 共享解析器
go test ./internal/cat -run 'TestRunFiles' -v
go test ./... -race                 # 竞态检测
go test ./... -cover                # 覆盖率
go test -run '^$' ./...             # 只编译测试，不运行（脚手架自检）
```

一键检查：

```powershell
pwsh -NoProfile -File scripts/check.ps1 -SkipTests   # 只做 gofmt/vet/build
pwsh -NoProfile -File scripts/check.ps1              # 全部检查
pwsh -NoProfile -File scripts/check.ps1 -Tool cat    # 只测 cat
```

注意：实现完成前，完整检查会在测试阶段失败，这是预期的。测试细节见 [docs/testing.md](docs/testing.md)。

## 规范边界

本仓库追求“像真实工具一样工作”，但**不保证逐字节 GNU 兼容**。所有行为以 `docs/specs/` 为准，例如：

- 只要求短选项和 `--`，不要求长选项。
- 不提供 `--help` / `--version`。
- `cat` 不要求 `-v` / `-A` / `-e` / `-t` / `-u`。
- `tail` 不要求 `-f`。
- `wc` 使用单空格分隔、不做 GNU 风格列宽对齐。

这些都在各 spec 的“进阶练习”中列出，作为你自行扩展和补测试的方向。

> 文档示例使用 `$` 开头的 POSIX shell 记法。在 PowerShell 中，可用 `Get-Content -Raw file | ...` 提供 stdin，或直接使用文件参数；示例里的 `testdata/` 路径相对于对应的 `internal/<tool>` 目录。

## 新增一个工具

1. 创建 `internal/<name>/<name>.go`，实现 `Run` 签名（可先放 `cli.NotImplemented`）。
2. 创建 `cmd/<name>/main.go`，照抄现有入口并替换包名。
3. 在 `internal/<name>/<name>_test.go` 写验收测试，需要固定输入时放到 `internal/<name>/testdata/`。
4. 写 `docs/specs/<name>.md`，包含测试对照表。
5. 更新 `README.md` 的学习顺序、`docs/progress.md` 和 `docs/specs/README.md`。

## 常见问题

**为什么一开始所有测试都失败？**
因为 `Run` 是存根，返回 `not implemented`。这正是 TDD 的起点。

**可以修改测试吗？**
可以修复测试中的笔误，但不能为了通过而放宽断言。测试与 spec 是这个练习的契约。

**为什么用法错误用退出码 2，而不是 1？**
这是本仓库的统一约定：2 表示“命令根本没有开始执行”，1 表示“执行了但发生错误”。见 [docs/architecture.md](docs/architecture.md)。

**Windows 会把换行转成 CRLF 吗？**
不会。工具按字节处理，`\n` 是行分隔符；测试用例通过内存或 `testdata` 提供固定字节。`.gitattributes` 保证夹具始终以 LF 检出。

**可以用标准库 `flag` 包吗？**
可以，但它不支持组合短选项、附加值和选项与操作数混排。要满足测试，通常需要手写扫描：要么实现 `internal/opt` 的 `Parse`
（推荐），要么在每个工具里各写一份。

**`go run ./cmd/cat file` 输出 `cat: not implemented` 正常吗？**
正常，说明 `cmd` 接线正确、实现尚未开始。

## 参考资料

- [GNU Coreutils 手册](https://www.gnu.org/software/coreutils/manual/coreutils.html)
- [POSIX 工具规范](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/)
- [Go `io` 包](https://pkg.go.dev/io)
- [Go `bufio` 包](https://pkg.go.dev/bufio)
