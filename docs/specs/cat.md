# cat 规范

## 1. 目标

按顺序读取一个或多个文件（或 stdin），把内容写到 stdout，并可选地添加行号、压缩空行、显示行尾与制表符。

## 2. 用法

```text
cat [-bnsET] [FILE]...
```

## 3. 选项

| 选项 | 含义 |
|---|---|
| `-n` | 给所有行编号 |
| `-b` | 只给非空行编号；**无论顺序如何都覆盖 `-n`** |
| `-s` | 把连续 2 个及以上的空行压缩为 1 个 |
| `-E` | 在每个换行符前显示 `$` |
| `-T` | 把 TAB 显示为 `^I` |

- 选项可组合、可出现在操作数之后，`--` 结束选项解析。
- Cat 没有带值选项。

## 4. 行为细节

### 4.1 输入来源

- 没有操作数时读取 stdin。
- 操作数 `-` 表示 stdin，可以和其他文件混用，按出现顺序读取。
- 操作数按顺序拼接成一个连续输出流。

### 4.2 行号（`-n` / `-b`）

- 行号从 1 开始，**跨多个文件连续**，不因换文件而重置。
- 格式为 `%6d\t`：十进制右对齐宽度 6，后跟一个 TAB。
- `-n`：所有行都编号，包括空行。
- `-b`：只给非空行编号；空行不输出行号，也**不增加行号计数**。
- 两者同时出现时 `-b` 胜出，与命令行顺序无关。

示例（输入 `a\n\nb\n`）：

```text
$ cat -n
     1	a
     2	
     3	b
$ cat -b
     1	a

     2	b
```

最后一行结尾的换行由输入决定；编号前缀不改变原始换行。

### 4.3 压缩空行（`-s`）

- “空行”指内容为空、只包含一个 `\n` 的行。
- 连续 2 个及以上空行压缩为 1 个空行。
- 压缩状态是**全局**的：跨文件边界时，前一个文件末尾的空行与后一个文件开头的空行属于同一段。
- 输入开头的连续空行也按同样规则压缩。
- 单个空行保持不变。

### 4.4 显示变换（`-E` / `-T`）

- `-E`：在每个 `\n` 之前立即写一个 `$`，例如 `a\n` 输出 `a$\n`。
  - 文件末尾没有 `\n` 时，不补 `$`，也不补换行。
- `-T`：把每个 TAB 字节（0x09）替换为两个字符 `^I`。
- 两个选项可组合；变换后的内容仍然保留原有行号前缀。
- 组合示例：`-nE` 对 `a\nb\n` 输出：

```text
     1	a$
     2	b$
```

### 4.5 结尾换行与二进制内容

- 完全保留输入字节；不自动补结尾换行，也不删除已有换行。
- 最后一段没有 `\n` 的内容仍算一行：`-n` 会给它编号，但不添加换行。
- `\r` 是普通内容字节，CRLF 不会被转换。
- 允许透传 NUL 和其他非 UTF-8 字节。

### 4.6 多文件与错误

- 文件按操作数顺序读取。
- 某个文件无法打开或读取时：向 stderr 报错，继续处理后续文件，最终退出码为 1。
- 空文件不产生输出，也不算错误。

## 5. 退出码

| 退出码 | 场景 |
|---|---|
| 0 | 所有文件读取成功 |
| 1 | 至少一个文件读取失败（其余文件仍已处理） |
| 2 | 未知选项等用法错误 |

## 6. 错误信息

用法错误（精确文本）：

```text
cat: unrecognized option: -z
```

文件错误（前缀与路径固定，底层文本由操作系统决定）：

```text
cat: <路径>: <底层错误>
```

## 7. 示例

以下示例中的 `testdata/` 路径相对于 `internal/cat` 目录。

```text
$ cat testdata/one.txt
one
two
$ cat -n testdata/one.txt testdata/two.txt
     1	one
     2	two
     3	three
     4	four
$ cat -E testdata/one.txt
one$
two$
$ cat -T file-with-tabs.txt
a^Ib
$ cat -s blanks.txt
a

b
```

## 8. 测试对照

| 测试函数 | 子测试组 | 验证点 |
|---|---|---|
| `TestRun` | `stdin passthrough`、`empty stdin`、`single dash reads stdin` | 输入来源与字节透传 |
| `TestRun` | `number all lines`、`number non-blank lines*`、`combined -nE` | 行号、`-b` 覆盖、组合选项 |
| `TestRun` | `squeeze blank lines`、`squeeze leading blank lines` | `-s` 局部行为 |
| `TestRun` | `show ends*`、`show tabs` | `-E` / `-T` 与末尾无换行 |
| `TestRun` | `CRLF is preserved`、`numbered CRLF is preserved` | 二进制/换行保持 |
| `TestRun` | `double dash keeps reading stdin` | `--` |
| `TestRunFiles` | `single file`、`multiple files concatenated`、`empty file`、`no trailing newline preserved` | 文件输入 |
| `TestRunFiles` | `stdin operand among files` | `-` 混用与顺序 |
| `TestRunFiles` | `numbering continues across files` | 跨文件连续编号 |
| `TestRunFiles` | `squeeze state is global across files` | 跨文件压缩状态 |
| `TestRunFiles` | `options after operands`、`double dash ends options` | 解析顺序 |
| `TestRunErrors` | `unknown option` | 用法错误退出 2 |
| `TestRunErrors` | `missing file*` | 错误后继续、退出 1 |

## 9. 实现注意

- 推荐用 `bufio.Reader` 按字节或按行读取；行号与压缩状态需要在多个文件之间保持。
- `-b` 需要先判断一行是否为空，才能决定是否编号；可以按行读取或缓冲当前行。
- `-s` 的“空行”是内容为空的行，不是“连续 `\n` 字节”的简单计数；注意开头与文件边界。
- 行号前缀、`-E`、`-T` 的叠加顺序是：先输出行号前缀，再输出经过显示变换的行内容。
- 自己打开的文件用 `defer Close`；不要关闭 stdin。
- stdout 写失败时建议返回 1（不纳入验收测试）。

## 10. 进阶练习

- `-v`：用 `^` / `M-` 表示不可打印字节。
- `-A`（等价于 `-vET`）、`-e`（`-vE`）、`-t`（`-vT`）。
- `-u`：无缓冲输出。
- 严格流式处理超大文件，并补压力测试。

## 11. 参考

- [GNU cat 手册](https://www.gnu.org/software/coreutils/manual/html_node/cat-invocation.html)
- [POSIX cat](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/cat.html)
