# head 规范

## 1. 目标

输出每个输入的开头部分：前 N 行、前 N 字节，或默认前 10 行。

## 2. 用法

```text
head [-n N] [-c N] [-qv] [FILE]...
```

## 3. 选项

| 选项 | 含义 | 默认 |
|---|---|---|
| `-n N` | 输出前 N 行 | N = 10 |
| `-c N` | 输出前 N 字节 | — |
| `-q` | 不显示文件头 | — |
| `-v` | 总是显示文件头 | — |

- `-n` 和 `-c` 都必须是 `>= 0` 的十进制整数；可以使用 `-n5` 或 `-n 5`。
- `-n` / `-c` 同时出现时，**最后一个生效**（`-n 1 -c 3` 表示按字节取 3）。
- `-q` / `-v` 同时出现时，**最后一个生效**。
- 不支持负数计数；`-n -1` 会被当作 `-n` 的值 `-1`，属于非法数字（退出 2）。
- 选项可组合（如 `-qn2` 表示 `-q` 加 `-n 2`），可出现在操作数之后，`--` 结束选项解析。
- 不带值的未知选项 → 退出 2。

## 4. 行为细节

### 4.1 输入来源

- 没有操作数时读取 stdin。
- 操作数 `-` 表示 stdin；可与其他文件混用，按顺序处理。
- 逐个打开文件，多个文件依次输出。

### 4.2 选择规则

- 按行（`-n`）：
  - 行以 `\n` 分隔；最后一段没有 `\n` 的内容也算一行。
  - 输出恰好前 N 行；输入不足 N 行时输出全部。
  - 不自动补结尾换行：最后一行原本没有 `\n`，输出也没有。
  - N = 0 时不输出内容。
- 按字节（`-c`）：
  - 输出前 N 字节；输入不足时输出全部。
  - N = 0 时不输出内容。
  - 可能从多字节字符中间截断，这是预期行为。
- 解析完成后再开始 I/O；参数非法时不读取任何文件。

### 4.3 文件头

- 默认规则：
  - 操作数数量 > 1 时显示文件头；
  - 0 或 1 个操作数时不显示。
- `-q` 强制关闭，`-v` 强制开启；以最后出现的为准。
- 格式：`==> <名称> <==`，单独一行。
- 名称：
  - 普通文件使用命令行中写的路径文本；
  - 隐式 stdin 和显式 `-` 都显示为 `standard input`。
- 分隔：
  - 第一个**成功打开**的输入前不加空行；
  - 之后每个成功打开的输入前加一个空行；
  - 打开失败的文件不输出文件头。
- 成功打开的空文件仍然输出文件头（如果有文件头模式）。

示例（`head -n 2 a.txt b.txt`）：

```text
==> a.txt <==
1
2

==> b.txt <==
x
```

### 4.4 多文件与错误

- 某个文件无法打开时：写 stderr 错误，跳过它，继续处理后续文件，最终退出 1。
- 其余文件的输出和文件头不受影响。
- 文件头模式由操作数数量和 `-q`/`-v` 决定，不因某个文件失败而改变。

## 5. 退出码

| 退出码 | 场景 |
|---|---|
| 0 | 全部成功 |
| 1 | 至少一个文件失败（其余文件仍处理） |
| 2 | 未知选项、缺少值、非法数字 |

## 6. 错误信息

用法错误（精确文本）：

```text
head: unrecognized option: -z
head: option requires an argument: -n
head: invalid number: abc
head: invalid number: -1
```

文件错误（前缀与路径固定）：

```text
head: <路径>: <底层错误>
```

## 7. 示例

以下示例中的 `testdata/` 路径相对于 `internal/head` 目录；`example.txt` 的内容假设为 `abcdef\n`。

```text
$ head -n 2 testdata/a.txt
1
2
$ head -c 3 example.txt
abc
$ head -n 1 testdata/a.txt testdata/b.txt
==> testdata/a.txt <==
1

==> testdata/b.txt <==
x
$ head -q -n 1 testdata/a.txt testdata/b.txt
1
x
```

（PowerShell 中可以使用 `Get-Content -Raw` 提供 stdin 内容。）

## 8. 测试对照

| 测试函数 | 子测试组 | 验证点 |
|---|---|---|
| `TestRun` | `default is ten lines`、`count larger than input` | 默认值与短输入 |
| `TestRun` | `line count *`、`zero lines` | `-n`、附加值、0 |
| `TestRun` | `byte count *`、`zero bytes` | `-c`、附加值、0 |
| `TestRun` | `no trailing newline preserved`、`line without trailing newline` | 末尾无换行 |
| `TestRun` | `last of line and byte count wins`、`last of byte and line count wins` | 覆盖顺序 |
| `TestRun` | `double dash ends options`、`dash operand reads stdin` | 解析边界 |
| `TestRun` | `verbose shows stdin header` | stdin 名称 |
| `TestRunFiles` | `multiple files with headers`、`quiet suppresses headers`、`verbose adds header for one file` | 文件头规则 |
| `TestRunFiles` | `later quiet beats verbose`、`later verbose beats quiet` | `-q`/`-v` 顺序 |
| `TestRunFiles` | `verbose empty file still shows header` | 空文件文件头 |
| `TestRunFiles` | `stdin operand among files` | stdin 混用 |
| `TestRunFiles` | `options after operands`、`combined quiet and count` | 混排与选项簇 |
| `TestRunErrors` | `invalid number`、`negative number is unsupported`、`missing option value` | 用法错误 |
| `TestRunErrors` | `bad argument stops before reading files` | 解析先于 I/O |
| `TestRunErrors` | `missing file is reported and skipped`、`single missing file` | 错误继续与退出码 |

## 9. 实现注意

- 先完整解析参数，再打开文件；否则非法数字的测试会读到文件。
- 按行读取时可使用 `bufio.Reader.ReadString('\n')` 或逐字节扫描；注意最后没有换行的行。
- 按字节读取时，读满 N 字节即可停止；不要为了输出而多读到下一行。
- 文件头只在成功打开后打印；用一个布尔值记录“是否已经输出过文件头块”。
- 多个文件之间不要复用未重置的缓冲区内容。
- 关闭自己打开的文件，不关闭 stdin。

## 10. 进阶练习

- 支持负数：`-n -N` 输出除最后 N 行以外的全部内容；`-c -N` 同理。
- 长选项 `--lines`、`--bytes`、`--quiet`、`--verbose`。
- `-z`：以 NUL 分隔记录。
- 为文件头、负数、长选项补充测试。

## 11. 参考

- [GNU head 手册](https://www.gnu.org/software/coreutils/manual/html_node/head-invocation.html)
- [POSIX head](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/head.html)
