# tail 规范

## 1. 目标

输出每个输入的末尾部分：最后 N 行、从第 N 行开始的内容，或最后 N 字节。

## 2. 用法

```text
tail [-n N] [-n +N] [-c N] [-qv] [FILE]...
```

## 3. 选项

| 选项 | 含义 | 默认 |
|---|---|---|
| `-n N` | 输出最后 N 行 | N = 10 |
| `-n +N` | 从第 N 行开始输出（N ≥ 1） | — |
| `-c N` | 输出最后 N 字节 | — |
| `-q` | 不显示文件头 | — |
| `-v` | 总是显示文件头 | — |

- 普通 `N` 和 `-c N` 的 N 必须是 `>= 0` 的十进制整数；`-n +N` 的 N 必须 `>= 1`。
- `-n +0`、负数、非十进制文本都视为非法数字（退出 2）。
- `-n` / `-c` 同时出现时，**最后一个生效**；`-n +N` 按 `-n` 计算。
- `-q` / `-v` 最后一个生效。
- 支持附加值（`-n2`、`-c2`、`-n+2`）与选项簇（如 `-qn2`）。
- 选项可出现在操作数之后，`--` 结束选项解析。
- **不支持 `-f`**（进阶练习）。

## 4. 行为细节

### 4.1 输入来源

- 没有操作数时读取 stdin。
- 操作数 `-` 表示 stdin，可与其他文件混用。
- 多个文件按顺序输出。

### 4.2 选择规则

- 最后 N 行（`-n N`）：
  - N = 0 时不输出内容。
  - 输入行数少于 N 时输出全部。
- 从第 N 行开始（`-n +N`）：
  - 行号从 1 开始；N = 1 等价于输出全部。
  - N 超过总行数时输出为空。
  - 空输入始终输出为空。
- 最后 N 字节（`-c N`）：
  - N = 0 时不输出内容。
  - 输入字节数少于 N 时输出全部。
  - 可能从多字节字符中间开始，这是预期行为。
- 行定义为以 `\n` 分隔；最后一段没有 `\n` 的内容也算一行。
- 不自动补结尾换行；输入最后一行没有换行时，输出也没有。
- 解析完成后再开始 I/O；参数非法时不读取任何文件。

### 4.3 文件头

规则与 head 完全一致：

- 操作数数量 > 1 时默认显示文件头；0 或 1 个操作数时不显示。
- `-q` 关闭、`-v` 开启，以最后出现的为准。
- 格式：`==> <名称> <==`。
- 名称：普通文件使用命令行路径；隐式 stdin 和显式 `-` 都显示为 `standard input`。
- 第一个成功打开的输入前不加空行，之后每个成功打开的输入前加一个空行。
- 打开失败的文件不输出文件头；成功打开的空文件仍输出文件头。

### 4.4 多文件与错误

- 某个文件无法打开时：写 stderr 错误，跳过它，继续处理后续文件，最终退出 1。
- 其余文件照常输出。
- 文件头模式由操作数数量和 `-q`/`-v` 决定，不因失败而改变。

## 5. 退出码

| 退出码 | 场景 |
|---|---|
| 0 | 全部成功 |
| 1 | 至少一个文件失败（其余文件仍处理） |
| 2 | 未知选项、缺少值、非法数字 |

## 6. 错误信息

用法错误（精确文本）：

```text
tail: unrecognized option: -z
tail: option requires an argument: -n
tail: invalid number: abc
tail: invalid number: +0
tail: invalid number: -1
```

文件错误（前缀与路径固定）：

```text
tail: <路径>: <底层错误>
```

## 7. 示例

以下示例中的 `testdata/` 路径相对于 `internal/tail` 目录。

```text
$ tail -n 2 testdata/a.txt
4
5
$ tail -n +4 testdata/a.txt
4
5
$ tail -c 2 testdata/a.txt
5
$ tail -n 1 testdata/a.txt testdata/b.txt
==> testdata/a.txt <==
5

==> testdata/b.txt <==
y
```

（`testdata/a.txt` 为 `1\n2\n3\n4\n5\n`，`b.txt` 为 `x\ny\n`。）

## 8. 测试对照

| 测试函数 | 子测试组 | 验证点 |
|---|---|---|
| `TestRun` | `default is ten lines`、`last N lines`、`line count attached`、`count larger than input` | 默认值与 `-n` |
| `TestRun` | `zero lines` | N = 0 |
| `TestRun` | `from line N`、`from line one`、`from line N without trailing newline` | `-n +N` |
| `TestRun` | `byte count`、`byte count without trailing newline` | `-c` |
| `TestRun` | `no trailing newline preserved` | 不补换行 |
| `TestRun` | `last of line and byte count wins`、`last of byte and line count wins` | 覆盖顺序 |
| `TestRun` | `double dash ends options`、`dash operand reads stdin`、`verbose shows stdin header` | 解析与 stdin |
| `TestRun` | `CRLF line endings are preserved` | 字节保持 |
| `TestRunFiles` | `multiple files with headers`、`quiet suppresses headers`、`verbose adds header for one file` | 文件头 |
| `TestRunFiles` | `later quiet beats verbose`、`later verbose beats quiet` | `-q`/`-v` 顺序 |
| `TestRunFiles` | `verbose empty file still shows header` | 空文件文件头 |
| `TestRunFiles` | `stdin operand among files`、`options after operands` | 混用与混排 |
| `TestRunFiles` | `no trailing newline file` | 文件末尾无换行 |
| `TestRunErrors` | `invalid number`、`plus zero is invalid`、`negative number is invalid`、`missing option value` | 用法错误 |
| `TestRunErrors` | `bad argument stops before reading files` | 解析先于 I/O |
| `TestRunErrors` | `missing file is reported and skipped`、`single missing file` | 错误继续 |

## 9. 实现注意

- 从 stdin 无法 seek，普通 `-n N` 需要环形缓冲或整体读取；从文件可以 seek，但统一缓冲实现更简单。
- 实现 `-n +N` 时要正确区分“从第 N 行开始”和“最后 N 行”。
- 最后一行没有 `\n` 时要保留原状；不要在输出末尾补换行。
- 文件头逻辑与 head 相同；建议先抽出一个内部函数，避免两处行为漂移。
- 关闭自己打开的文件，不关闭 stdin。

## 10. 进阶练习

- `-f`：持续跟踪文件追加内容；实现后设计可控的测试（写入、等待、停止）。
- `-F`：文件被替换/轮转后继续跟踪。
- `-s`、`--pid`、`-c +N`、长选项。
- 用环形缓冲实现真正的流式 `-n N`，并补大文件测试。

## 11. 参考

- [GNU tail 手册](https://www.gnu.org/software/coreutils/manual/html_node/tail-invocation.html)
- [POSIX tail](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/tail.html)
