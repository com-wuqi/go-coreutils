# 架构与统一约定

本文描述所有工具共享的约定。单个工具的参数与行为以 `docs/specs/<tool>.md` 为准。

## 1. 分层

```text
cmd/<tool>/main.go
      │  只做接线：os.Args、os.Stdin、os.Stdout、os.Stderr
      ▼
internal/<tool>/<tool>.go
      │  全部命令逻辑；依赖标准库和 internal 下的共享包
      ▼
internal/cli
         退出码、RunFunc、NotImplemented（无命令逻辑）
internal/opt
         共享短选项扫描（Parse 待实现，见 5.4）

internal/<tool>/<tool>_test.go
      │  只调用 <tool>.Run，不触碰内部实现
      ▼
internal/testutil
         捕获 stdout/stderr、断言、临时文件
```

约束：

- `cmd` 不做参数解析、不打印错误、不决定退出码。
- `internal/<tool>` **可以**导入 `internal/cli` 等纯共享包；**不要导入其他工具包**，避免 `cat` 依赖 `head` 这类工具间耦合。
- 共享能力放在 `internal/<shared>` 下（例如 `cli`、`testutil`，以及后续可抽取的参数解析器）；共享包不得包含单个命令的业务策略或文件 I/O。
- `internal/cli` 只放跨工具约定，不放某个工具的行为。
- `internal/testutil` 仅用于测试；命令行二进制不会导入它。

> Go 的 `internal/` 限制的是**模块外**导入；在本模块内，工具包导入任意共享内部包是正常用法。真正要避免的是工具包之间的横向依赖，以及把单个命令的策略塞进共享包。

## 2. `Run` 契约

```go
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int
```

- `args` 是程序名之后的参数，可能为空。
- 读取输入只能来自 `stdin` 和 `args` 指向的文件。
- 正常输出只写 `stdout`；错误与用法信息只写 `stderr`。
- 返回退出码；不调用 `os.Exit`、`log.Fatal`、`panic`。
- 不关闭 `stdin`、`stdout`、`stderr`；自己打开的文件必须关闭。
- 不读取全局变量（`os.Args`、环境变量、全局状态）；保证同一进程内可重复调用。
- 允许内部使用缓冲；对用户可见的行为必须与规范一致。

`cmd/<tool>/main.go` 固定为：

```go
func main() {
	os.Exit(tool.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
```

## 3. 退出码

| 值 | 名称 | 使用场景 |
|---|---|---|
| 0 | `cli.ExitOK` | 全部操作成功 |
| 1 | `cli.ExitFailure` | 执行中发生运行时错误；能继续的输入继续处理，最终返回 1 |
| 2 | `cli.ExitUsage` | 参数无法解析；不执行任何 I/O，不产生 stdout |

要点：

- 用法错误先于一切 I/O。即使操作数中包含可读文件，也直接返回 2。
- 多文件命令遇到坏文件时继续处理其余文件，最终返回 1。
- 退出码 1 不要求 stderr 恰好一行，但每个失败输入至少有一行错误信息。

## 4. 错误信息

用法错误（文本由本仓库固定，测试精确匹配）：

```text
<prog>: unrecognized option: <完整原始参数>
<prog>: option requires an argument: -<选项字符>
<prog>: invalid number: <原始值>
```

文件错误（底层错误文本由操作系统决定，测试只匹配前缀与路径）：

```text
<prog>: <路径>: <底层错误>
```

规则：

- 每行以 `\n` 结尾。
- 不使用 `log` 包（它会加时间戳）。
- 不输出 usage 长文本；进阶功能可以自行添加 `--help`。
- 程序名常量与目录名一致：`cat`、`echo`、`head`、`tail`、`tee`、`wc`。

## 5. 参数解析

适用范围：cat、head、tail、wc、tee。echo 使用第 5.3 节的特殊规则。

### 5.1 扫描规则

从 `args` 左到右扫描，先完成全部解析，再开始 I/O：

1. 参数等于 `--`：结束选项解析，其余全部是操作数。
2. 参数等于 `-`：操作数（对 cat/head/tail/wc 表示 stdin；对 tee 是文件名）。
3. 参数以 `-` 开头且长度大于 1：按短选项簇解析，例如 `-nE`。
4. 其他参数：操作数，保持出现顺序。
5. 选项和操作数可以混排；`cat file -n` 与 `cat -n file` 等价。

### 5.2 短选项簇与值

- 无值选项逐字符处理，例如 `-nET`。
- 带值选项：
  - 若簇中该字符后还有字符，剩余部分作为值，例如 `-n5`、`-qn5`。
  - 否则消耗下一个参数作为值，即使它以 `-` 开头，例如 `tail -n +2`、`head -n -1`。
  - 没有下一个参数 → `<prog>: option requires an argument: -n`，退出 2。
- 同一个选项重复出现时，通常“最后一次生效”；具体覆盖规则见各 spec（例如 cat 的 `-b` 永远覆盖 `-n`）。
- 未识别的长选项（如 `--lines=5`）按未知选项处理：`<prog>: unrecognized option: --lines=5`。
- 非法数字（空、非十进制、负数而规范不允许）→ `<prog>: invalid number: <原值>`，退出 2。

### 5.3 echo 的特殊规则

echo 没有通用扫描器：

- 只处理**前置**参数；只要参数完整匹配 `-[neE]+` 就是选项簇，可组合。
- 选项按出现顺序生效（`-e -E` 关闭转义，`-E -e` 开启转义）。
- 第一个不匹配的参数开始输出；它和之后的所有参数都按字面量处理，包括 `--`、`-z`、`-nx`。

### 5.4 共享解析器（脚手架已建立，`Parse` 待实现）

> 状态：类型、`UsageError`、存根、25 个验收测试和 [opt.md](specs/opt.md) 已就绪；`Parse` 仍返回 `opt: not implemented`
> ，五个工具尚未接入。

**目标**：把 5.1/5.2 的纯语法扫描抽到 `internal/opt`，让 cat/head/tail/wc/tee 复用；echo 不接入。

**包与 API**（实现见 `internal/opt/opt.go`）：

```go
// internal/opt/opt.go（Parse 为存根）
package opt

// Option 声明一个短选项。TakesValue 为 true 时该选项需要一个值。
type Option struct {
	Short      byte
	TakesValue bool
}

// Occurrence 是命令行中出现的一次选项。
type Occurrence struct {
	Short byte
	Value string // 无值选项时为空
}

// Result 是解析结果。
type Result struct {
	Operands []string     // 保持命令行顺序
	Options  []Occurrence // 保持命令行顺序，供工具实现 last-wins
}

// UsageError 表示参数语法错误；Error() 返回 "<prog>: <msg>"。
type UsageError struct {
	Prog string
	Msg  string
}

// Parse 按 options 扫描 args。它只做语法层，不校验数字、不处理 I/O。
func Parse(prog string, options []Option, args []string) (Result, error)
```

语法错误只覆盖两类，消息格式与第 4 节一致：

- `unrecognized option: <完整原始 token>`，例如 `unrecognized option: -qz`；
- `option requires an argument: -<选项字符>`，例如 `option requires an argument: -n`。

工具的使用方式（示意）：

```go
parsed, err := opt.Parse("head", []opt.Option{
	{Short: 'n', TakesValue: true},
	{Short: 'c', TakesValue: true},
	{Short: 'q'},
	{Short: 'v'},
}, args)
if err != nil {
	fmt.Fprintln(stderr, err)
	return cli.ExitUsage
}
for _, o := range parsed.Options {
	switch o.Short {
	case 'n':
		// 在这里做数字校验和 last-wins；非法值由工具输出 invalid number。
	case 'q':
		// ...
	}
}
```

**职责边界**：

| 层 | 负责 |
|---|---|
| `internal/opt` | `--`、单独 `-`、短选项簇、附加值/下一参数值、选项与操作数混排、未知选项与缺值错误、保留出现顺序 |
| 各工具 | 数字/取值校验（`invalid number`）、last-wins 与覆盖规则（如 `-b` 覆盖 `-n`）、文件 I/O、文件头、计数、运行时错误、退出码 |
| echo | 完全保留 5.3 的特殊解析，不依赖 `internal/opt` |

**落地状态**：

1. [x] 新增 `internal/opt/opt.go`（类型 + `Parse` 存根）与 `internal/opt/opt_test.go`（25 个用例）。
2. [x] 新增 `docs/specs/opt.md`，并更新 `docs/specs/README.md` 与 `docs/progress.md`。
3. [ ] 实现 `Parse`，让 `go test ./internal/opt` 全绿。
4. [ ] 五个工具改为调用 `opt.Parse`；现有工具验收测试保持不变（它们只测 `Run` 的行为）。
5. [ ] 删除本节和 `Parse` 中的“待实现”标记。

**备选方案**：继续让每个工具自行解析。优点是解析规则是重要练习内容；缺点是同一套边界逻辑重复五遍，修一处要改五处。当前已选择共享方案，但工具仍可自行解析——工具验收测试只检查行为，不检查是否调用
`internal/opt`。

## 6. 字节、换行与流

- 所有输入输出按字节处理，不做字符集转换或 CRLF 转换。
- `\n` 是唯一行分隔符；`\r` 是普通内容字节。
- 不自动补结尾换行，也不删除输入中已有的换行。
- 文件末尾没有 `\n` 时，最后一段内容仍算“一行”，输出时保持没有换行。
- 二进制数据（含 NUL、0xFF）必须可透传。
- 不要假设文件很小。练习允许先整体读入，但 cat/head/tee 更推荐流式读取。

## 7. 固定输出格式

### 7.1 cat 行号

`fmt` 语义等价于 `%6d\t`：右对齐宽度 6，后跟一个 TAB。

```text
     1	first
     2	second
```

### 7.2 head/tail 文件头

```text
==> <名称> <==
```

- 名称使用命令行中的路径文本；隐式 stdin 和显式 `-` 都显示为 `standard input`。
- 第一个成功打开的文件前不加空行；之后每个成功打开的文件前加一个空行。
- 文件打开失败时不输出文件头，直接报错并继续。
- 空文件成功打开时仍输出文件头。

### 7.3 wc 列

- 固定顺序：`lines words chars bytes max-line-length`。
- 只输出被请求的列；默认 `lines words bytes`。
- 字段之间一个空格；数字不补齐、不对齐。
- 有文件名时，数字之后追加一个空格和文件名。
- 多操作数总计行的名称固定为 `total`。

## 8. 建议的实现边界

- 计数、行号使用 `int64`，避免大文件溢出。
- 只使用标准库即可完成全部练习；`bufio`、`bytes`、`io`、`os`、`strconv`、`strings` 足够。
- 并发不是必需的。若使用 goroutine，必须自行保证结果顺序稳定、无数据竞争。
- 文件打开后立即用 `defer f.Close()` 或集中关闭，避免泄漏。
- 输出错误（stdout 关闭、磁盘满）建议返回 1，但测试无法稳定模拟，不纳入验收。

## 9. 测试边界

- 验收测试只通过 `Run` 观察行为，不读取未导出字段或内部函数。
- `testutil.RunCases` 会断言 stdout、stderr、退出码三者。
- 文件系统错误只做子串断言，避免依赖 Windows/中文系统的错误文本。
- 测试不构建二进制、不依赖网络、不修改当前工作目录。
