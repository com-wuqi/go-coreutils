# wc 规范

## 1. 目标

统计输入的行数、词数、字符数、字节数和最长行长度。

## 2. 用法

```text
wc [-lwcmL] [FILE]...
```

## 3. 选项

| 选项 | 统计项 | 输出列名（文档用） |
|---|---|---|
| `-l` | 行数 | lines |
| `-w` | 词数 | words |
| `-m` | 字符数（Unicode code points） | chars |
| `-c` | 字节数 | bytes |
| `-L` | 最长行的字符数 | max-line-length |

- 没有任何统计选项时，默认等价于 `-l -w -c`。
- 选项可组合、可重复，重复无额外效果。
- 输出列顺序固定，与命令行中的选项顺序无关。
- 选项可出现在操作数之后，`--` 结束选项解析；没有带值选项。
- 未知选项 → 退出 2。

## 4. 行为细节

### 4.1 计数定义

- **lines**：输入中 `\n` 字节的数量。最后一段没有 `\n` 的内容不计入行数。
- **words**：由 ASCII 空白分隔的最大连续非空白片段数量。空白字节为：
  - 空格 `0x20`、`\t`、`\n`、`\v`、`\f`、`\r`。
- **bytes**：输入的总字节数。
- **chars**：输入按 UTF-8 解码后的 Unicode code point 数量，包含换行；非法字节序列中的每个非法字节计 1。
- **max-line-length**：所有行中“去除行尾 `\n` 后的 rune 数”的最大值。
  - `\r` 是普通字符，计入长度。
  - 最后一段没有 `\n` 的内容也参与计算。
  - 空输入或没有内容时为 0。

示例：

```text
输入 "hello world\nfoo bar baz\n"
lines = 2, words = 5, bytes = 24, chars = 24, max-line-length = 11

输入 "你好 世界\n"
lines = 1, words = 2, bytes = 14, chars = 6, max-line-length = 5
```

### 4.2 输出格式

- 字段之间用**一个空格**分隔，数字不做补齐或对齐。
- 只输出被请求的列，顺序固定为：`lines words chars bytes max-line-length`。
- 有文件名时，最后一个数字后接一个空格和文件名。
- 无操作数时读取 stdin，**不显示文件名**。
- 操作数 `-` 表示 stdin，文件名显示为 `-`。
- 普通文件显示命令行中的路径文本。
- 多操作数时，最后追加一行汇总，名称固定为 `total`。

示例：

```text
$ wc testdata/sample.txt
2 5 24 testdata/sample.txt
$ wc -l testdata/sample.txt testdata/unicode.txt
2 testdata/sample.txt
1 testdata/unicode.txt
3 total
$ printf 'abc\n' | wc -l
1
```

### 4.3 多文件与汇总

- 按操作数顺序逐个读取，每个成功读取的操作数输出一行。
- 操作数数量 >= 2 时输出 `total` 行：
  - 只汇总成功读取的文件；
  - 某个文件失败不影响 total；
  - 即使最终只有 1 个文件成功（另一个失败），仍输出 total。
- 所有文件都失败时，stdout 为空，退出 1。
- 单个文件失败时不输出 total 行。

### 4.4 错误

- 某个文件无法打开或读取：写 stderr 错误，继续处理后续操作数，最终退出 1。
- 参数错误：退出 2，不读取任何文件。

## 5. 退出码

| 退出码 | 场景 |
|---|---|
| 0 | 全部操作数读取成功 |
| 1 | 至少一个文件读取失败 |
| 2 | 未知选项等用法错误 |

## 6. 错误信息

用法错误（精确文本）：

```text
wc: unrecognized option: -x
```

文件错误（前缀与路径固定）：

```text
wc: <路径>: <底层错误>
```

## 7. 示例

以下示例中的 `testdata/` 路径相对于 `internal/wc` 目录。

```text
$ wc testdata/sample.txt
2 5 24 testdata/sample.txt
$ wc -lw testdata/sample.txt
2 5 testdata/sample.txt
$ wc -m testdata/unicode.txt
6 testdata/unicode.txt
$ wc -L testdata/sample.txt
11 testdata/sample.txt
$ printf 'a\n' | wc -l -
1 -
```

## 8. 测试对照

| 测试函数 | 子测试组 | 验证点 |
|---|---|---|
| `TestRun` | `default counts`、`empty stdin` | 默认三列、空输入 |
| `TestRun` | `no trailing newline` | 行/词/字节定义 |
| `TestRun` | `lines only`、`words only`、`bytes only`、`chars only`、`longest line only` | 单列选项 |
| `TestRun` | `columns keep fixed order`、`all columns keep fixed order` | 固定列顺序 |
| `TestRun` | `unicode counts` | rune、字节、最长行 |
| `TestRun` | `whitespace splits words`、`CRLF bytes count literally` | 分词与字节 |
| `TestRun` | `double dash ends options`、`dash operand is named`、`duplicate flags are harmless` | 解析与命名 |
| `TestRunFiles` | `single file`、`selected columns`、`empty file`、`no trailing newline file` | 文件输入 |
| `TestRunFiles` | `two files and total`、`stdin operand among files` | 汇总与 stdin |
| `TestRunFiles` | `options after operands`、`double dash ends options` | 混排 |
| `TestRunErrors` | `unknown option` | 用法错误 |
| `TestRunErrors` | `missing file before valid file`、`single missing file` | 错误继续、totals |

## 9. 实现注意

- 计数使用 `int64`；一次遍历同时维护多组计数器。
- 推荐按块读取（如 32 KiB），避免把整个文件读进内存。
- words 的空白判定使用固定字节集合，不要用 `unicode.IsSpace`（规范按 ASCII 空白）。
- chars 与 max-line-length 需要 UTF-8 解码：可用 `utf8.DecodeRune`，记录 `size` 并正确跳过非法字节。
- 最长行统计要按行重置当前长度，遇到 `\n` 时更新最大值。
- 输出名与汇总名要分开处理：无操作数时没有名字；显式 `-` 的名字是 `-`。
- 汇总行在读取结束后统一输出，避免边读边改 stdout。

## 10. 进阶练习

- GNU 风格列宽对齐（所有行按同一宽度填充）。
- `--files0-from`、长选项 `--lines` 等。
- `-L` 按终端显示宽度而不是 rune 数。
- 读取超大文件时的内存与速度基准测试。

## 11. 参考

- [GNU wc 手册](https://www.gnu.org/software/coreutils/manual/html_node/wc-invocation.html)
- [POSIX wc](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/wc.html)
