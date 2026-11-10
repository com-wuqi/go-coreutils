# 测试指南

本仓库采用“规范先行 + 验收测试”的方式：测试是 `Run` 公开行为的可执行契约。

## 1. 测试放在哪

- 每个工具一个外部测试包：`internal/<tool>/<tool>_test.go`，包名为 `<tool>_test`。
- 测试只导入 `go-coreutils/internal/<tool>` 和 `go-coreutils/internal/testutil`。
- 固定输入放在 `internal/<tool>/testdata/`；可写文件使用 `t.TempDir()`。
- 不为 `cmd` 写测试：入口只有三行接线。

## 2. 运行测试

```powershell
go test ./...                          # 全部
go test ./internal/echo -v             # 单工具，显示子测试
go test ./internal/cat -run TestRun    # 单个测试函数
go test ./internal/head -run 'TestRunFiles/multiple' -v
go test ./... -race                    # 竞态检测
go test ./... -cover                   # 覆盖率
go test -run '^$' ./...                # 只编译，不运行
```

`go test -run '^$' ./...` 是脚手架自检命令：它证明所有测试文件能编译。实现尚未开始时它应当通过，而普通 `go test ./...` 应当失败。

## 3. 初始红态

所有失败最终都源于 `cli.NotImplemented`（存根向 stderr 写错误并返回 1）：

```text
exit code mismatch:
     got: 1
    want: 0
```

这是预期状态。判断进展的方法：

- 失败数量减少 = 实现推进。
- 出现 `want: 2` 的失败 = 用法错误路径尚未实现。
- 出现 stdout 差异 = 核心行为尚未正确。

不要在实现前修改期望值来消除红态。

## 4. 测试辅助包

`internal/testutil/testutil.go` 提供：

```go
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, fn cli.RunFunc, stdin string, args ...string) Result

type Case struct {
	Name          string
	Stdin         string
	Args          []string
	WantStdout    string
	WantStderr    string
	WantStderrSub string
	WantExit      int
}

func RunCases(t *testing.T, fn cli.RunFunc, cases []Case)
func AssertEqual[T comparable](t *testing.T, label string, got, want T)
func AssertContains(t *testing.T, label, got, sub string)
func WriteFile(t *testing.T, dir, name, content string) string
func WriteFiles(t *testing.T, files map[string]string) (dir string)
```

用法示例：

```go
func TestRun(t *testing.T) {
	tests := []testutil.Case{
		{Name: "empty input", Stdin: "", WantStdout: ""},
		{Name: "unknown option", Args: []string{"-z"}, WantExit: 2,
			WantStderr: "echo: unrecognized option: -z\n"},
	}
	testutil.RunCases(t, echo.Run, tests)
}
```

`RunCases` 的断言规则：

- 退出码必须等于 `WantExit`（缺省 0）。
- stdout 必须与 `WantStdout` 完全相等。
- `WantStderr` 非空：stderr 精确匹配，用于本仓库固定文本（用法错误）。
- `WantStderrSub` 非空：stderr 包含该子串，用于操作系统错误文本。
- 两者都为空：stderr 必须为空。

需要检查文件内容时（例如 tee），直接用 `testutil.Run` 获取结果，再配合 `os.ReadFile`：

```go
got := testutil.Run(t, tee.Run, "hello\n", path)
testutil.AssertEqual(t, "stdout", got.Stdout, "hello\n")
data, err := os.ReadFile(path)
if err != nil {
	t.Fatalf("read %s: %v", path, err)
}
testutil.AssertEqual(t, "file content", string(data), "hello\n")
```

## 5. 编写新测试

1. 在对应 spec 中写清行为，并补一行“测试对照”。
2. 在 `_test.go` 中新增 `Case`，名字用能读懂行为的英文短语。
3. 优先断言完整输出；不要只检查 `strings.Contains`，除非处理的是系统错误文本。
4. 使用 `t.TempDir()` 隔离可写文件；用 `filepath.Join` 构造路径。
5. 不要 `t.Chdir`、不要依赖测试执行顺序、不要修改仓库里的 testdata。
6. 不要写“跳过测试”来隐藏未实现的核心功能；进阶功能可以放在单独的练习里。

## 6. 失败信息怎么读

```text
--- FAIL: TestRun/number_all_lines (0.00s)
    testutil.go:56: stdout mismatch:
         got: "..."
        want: "     1\ta\n"
```

- `stdout mismatch`：输出内容或格式不符。
- `stderr must contain ...`：预期错误信息缺少路径或前缀。
- `exit code mismatch`：退出码不符（0/1/2）。
- 先修第一个失败用例；后面的失败常常是同一原因。

## 7. Windows 与换行

- 工具按字节处理，绝不把 `\n` 转成 `\r\n`，反之亦然。
- 仓库配置了 `core.autocrlf` 兼容的 `.gitattributes`：`**/testdata/**` 强制以 LF 检出。
- 新增夹具时请保持 LF；需要测试 CRLF 时，用内存字符串（例如 `"a\r\n"`）或 `t.TempDir()` 动态写入。
- 文件路径出现在输出中时，用同一变量构造期望值，例如：

```go
a := filepath.Join("testdata", "a.txt")
want := "==> " + a + " <==\n"
```

不要硬编码 `testdata/a.txt`，否则 Windows 反斜杠会导致失败。

## 8. 覆盖率与竞态

```powershell
go test ./internal/cat -cover
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
go test ./... -race
```

覆盖率用于发现遗漏分支，不要为了数字写无意义断言。竞态检查在默认的单 goroutine 实现下通常很快。

## 9. 一键检查

```powershell
pwsh -NoProfile -File scripts/check.ps1 -SkipTests   # 格式/vet/构建
pwsh -NoProfile -File scripts/check.ps1 -Tool cat    # 只测 cat
pwsh -NoProfile -File scripts/check.ps1              # 全部
```

实现完成前，完整检查会以测试失败结束，这是预期行为。`-SkipTests` 应当始终通过。
