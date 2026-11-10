# tee 规范

## 1. 目标

从 stdin 读取一次，把相同的字节同时写到 stdout 和一个或多个文件。

## 2. 用法

```text
tee [-a] [FILE]...
```

## 3. 选项

| 选项 | 含义 |
|---|---|
| `-a` | 追加到文件；默认截断 |
| `--` | 结束选项解析 |

- 没有其他选项；`-a` 重复出现无额外效果。
- 选项可以出现在操作数之后。
- `-` 是**字面文件名**，不表示 stdout；本工具没有“用 `-` 代表 stdout”的约定。

## 4. 行为细节

### 4.1 输入与输出

- 从 stdin 读取全部字节，原样写到 stdout 和每个目标文件。
- 没有任何文件操作数时，只把 stdin 复制到 stdout。
- 不做换行转换、不补换行、不做编码转换；NUL 和任意二进制字节都原样透传。

### 4.2 文件打开模式

- 默认：创建或截断，只写。
- `-a`：创建（若不存在）并以追加方式打开。
- 文件权限使用 0666（受系统 umask 影响）。
- 每个文件只打开一次，内容与 stdout 完全一致。

### 4.3 失败处理

- 某个目标文件无法打开或写入时：
  - 向 stderr 输出 `tee: <路径>: <底层错误>`；
  - 跳过该目标；
  - 继续写 stdout 和其他文件；
  - 全部处理完后退出码为 1。
- 只有一个目标且打开失败时，stdout 仍然收到完整输入。
- 没有失败时退出码为 0。
- stdout 写失败时建议返回 1（不纳入验收测试）。

## 5. 退出码

| 退出码 | 场景 |
|---|---|
| 0 | 所有目标写入成功（或没有文件目标） |
| 1 | 至少一个目标打开/写入失败 |
| 2 | 未知选项等用法错误 |

## 6. 错误信息

用法错误（精确文本）：

```text
tee: unrecognized option: -z
```

文件错误（前缀与路径固定）：

```text
tee: <路径>: <底层错误>
```

## 7. 示例

```text
$ echo hi | tee out.txt
hi
$ cat out.txt
hi
$ echo more | tee -a out.txt
more
$ cat out.txt
hi
more
$ echo x | tee missing-dir/out.txt good.txt
tee: missing-dir/out.txt: ...
x
$ cat good.txt
x
```

（最后一条退出码为 1，但 stdout 与 `good.txt` 仍然写入。）

## 8. 测试对照

| 测试函数 | 子测试组 | 验证点 |
|---|---|---|
| `TestRun` | `no files passes stdin through`、`empty stdin` | 无文件目标 |
| `TestRun` | `binary passthrough` | 二进制透传 |
| `TestRun` | `append flag without files`、`double dash with no files` | 选项解析 |
| `TestRunFiles` | `writes stdin to one file`、`writes stdin to multiple files` | 基本多路写出 |
| `TestRunFiles` | `truncates an existing file by default`、`append keeps existing content`、`append creates a missing file` | 截断/追加 |
| `TestRunFiles` | `empty stdin creates an empty file` | 空输入 |
| `TestRunFiles` | `options after operands` | 选项混排 |
| `TestRunFiles` | `failed target does not stop other outputs`、`failing only target still copies to stdout` | 失败降级与退出码 |
| `TestRunErrors` | `unknown option` | 用法错误 |

## 9. 实现注意

- 推荐先把所有目标文件打开（同时记录失败），再开始读取 stdin，这样能“先失败、后写入”。
- 每读一块就写 stdout 和所有成功打开的文件；`io.MultiWriter` 可以简化，但要自行确认失败目标的处理。
- 在循环结束后关闭所有文件；不要关闭 stdout。
- 若采用一次性 `io.Copy` + `io.MultiWriter`，注意某个 writer 失败会让 `MultiWriter` 提前停止，需要按目标分别处理。
- 文件打开失败不应阻止继续读取 stdin，否则 stdout 会缺内容。

## 10. 进阶练习

- `-i`：忽略中断信号（平台相关）。
- `-p`：诊断写往非阻塞输出时的错误。
- `--output-error=MODE` 与长选项。
- 补一个测试：目标路径父目录不存在时的错误与继续行为。

## 11. 参考

- [GNU tee 手册](https://www.gnu.org/software/coreutils/manual/html_node/tee-invocation.html)
- [POSIX tee](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/tee.html)
