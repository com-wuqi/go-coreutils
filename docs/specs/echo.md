# echo 规范

## 1. 目标

输出参数字符串，支持关闭结尾换行与反斜杠转义。不读取 stdin，不访问文件。

## 2. 用法

```text
echo [-neE] [ARG]...
```

## 3. 选项

| 选项 | 含义 |
|---|---|
| `-n` | 不输出结尾换行 |
| `-e` | 开启转义解释 |
| `-E` | 关闭转义解释（默认状态） |

- 只在**参数开头**识别选项，且整个参数必须匹配 `-[neE]+`。
- 同一参数内可组合，例如 `-ne`、`-en`、`-eE`。
- 选项按从左到右顺序生效，因此 `-e -E` 最终关闭转义，`-E -e` 最终开启转义。
- 第一个不匹配 `-[neE]+` 的参数开始被当作输出内容，之后所有参数（包括 `--`、`-z`、`-nx`）都按字面量输出。

## 4. 行为细节

### 4.1 参数拼接

- 所有待输出参数用单个空格连接。
- 空字符串参数仍然占一个位置：`echo ""` 输出一个换行；`echo a "" b` 输出 `a  b\n`（两个空格）。
- 无参数时输出一个换行。
- 结尾换行规则：
  - 默认输出一个 `\n`；
  - 有 `-n` 时不输出；
  - 遇到 `\c` 时立即停止，之后的参数、字符和结尾换行都不输出。

### 4.2 转义表（仅 `-e` 生效）

| 序列 | 输出 |
|---|---|
| `\\` | 反斜杠 `\` |
| `\a` | BEL，0x07 |
| `\b` | 退格，0x08 |
| `\c` | 立即停止输出（不补结尾换行） |
| `\e` | ESC，0x1B |
| `\f` | 换页，0x0C |
| `\n` | 换行，0x0A |
| `\r` | 回车，0x0D |
| `\t` | 制表符，0x09 |
| `\v` | 垂直制表，0x0B |
| `\0NNN` | NUL 或八进制字节：`\0` 后跟 0–3 位八进制数字 |
| `\xHH` | 十六进制字节：`\x` 后跟 1–2 位十六进制数字 |

补充规则：

- `\0101` 是八进制 101 = 十进制 65 = `A`。
- `\x41` 也是 `A`；只有一位时按一位取值（如 `\x4` = 0x04）。
- 未知转义保留反斜杠和字符本身：`\q` 输出 `\q`。
- 参数末尾孤立的 `\` 原样输出。
- 不在 `-e` 模式时，所有序列原样输出。
- 输出按字节处理，允许产生 NUL；UTF-8 内容原样透传。

## 5. 退出码

| 退出码 | 场景 |
|---|---|
| 0 | 正常输出（包括 `-n`、`\c`、空参数） |
| 1 | 写 stdout 失败（可选实现，不纳入验收测试） |

## 6. 错误信息

本工具没有用法错误路径。正常实现不应向 stderr 写任何内容。

## 7. 示例

```text
$ echo hello world
hello world
$ echo -n hi
hi
$ echo -e 'a\tb'
a<TAB>b
$ echo -e 'a\nb'
a
b
$ echo -e 'a\cb'
a
$ echo -e -E 'a\nb'
a\nb
$ echo -n x -e
x -e
```

（`<TAB>` 表示一个制表符；最后一行说明选项解析在第一个非选项参数处停止。）

## 8. 测试对照

| 测试函数 | 子测试组 | 验证点 |
|---|---|---|
| `TestRun` | `no args prints newline`、`single arg`、`multiple args joined by space`、`empty argument *` | 拼接、空格、空参数 |
| `TestRun` | `-n omits trailing newline`、`-n only` | 关闭结尾换行 |
| `TestRun` | `-e interprets *`、`\\c suppresses newline` | 完整转义表、停止输出 |
| `TestRun` | `-E disables escapes`、`-e followed by -E`、`-E followed by -e` | 状态顺序 |
| `TestRun` | `combined -ne` | 选项组合 |
| `TestRun` | `option parsing stops at first operand`、`unknown leading option is literal`、`double dash is literal`、`mixed flags are literal` | 前置选项边界 |

## 9. 实现注意

- 可以先把输出写入 `bytes.Buffer` 或 `[]byte`，最后一次性写 stdout；也可以边解析边写。
- 解析八进制时要限制最多 3 位，十六进制最多 2 位，避免越界或误解。
- `\c` 一旦出现，后续参数不再输出，且不再补换行。
- `-e` / `-E` 的状态是顺序的，不要只检查“是否存在 `-e`”。
- 不要使用 `fmt.Sprintf` 来拼接含 NUL 的内容；直接写字节更稳妥。

## 10. 进阶练习

- `--help` / `--version`。
- `\uHHHH` / `\UHHHHHHHH` Unicode 转义。
- 更多八进制形式与 `\1` 风格转义（不同 shell 行为不同，需自行定义并补测试）。
- 处理 stdout 写失败并返回 1。

## 11. 参考

- [GNU echo 手册](https://www.gnu.org/software/coreutils/manual/html_node/echo-invocation.html)
- [POSIX echo](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/echo.html)
