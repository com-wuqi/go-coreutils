# 实现进度

初始状态：脚手架、规范与测试已完成；**所有命令实现与共享解析器的 `Parse` 均未开始**，测试全红。

## 总览

| 工具 | 规范 | 验收测试 | 参数解析 | 核心功能 | 错误处理 | 测试全绿 | 手工冒烟 |
|---|---|---|---|---|---|---|---|
| echo | ✅ | ✅ 28 例 | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ |
| cat | ✅ | ✅ 30 例 | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ |
| tee | ✅ | ✅ 15 例 | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ |
| head | ✅ | ✅ 31 例 | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ |
| tail | ✅ | ✅ 34 例 | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ |
| wc | ✅ | ✅ 27 例 | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ |

> 用例数包含每个测试函数下的子测试：6 个工具 165 个 + 共享解析器 25 个，合计 190 个。

## internal/opt — [规范](specs/opt.md)

共享参数解析器；cat/head/tail/wc/tee 的解析部分推荐复用它。

- [x] 类型、`UsageError`、`Parse` 存根
- [x] 25 个验收测试与规范
- [ ] 选项簇、附加值、独立值、`--`、`-`、混排
- [ ] 未知选项与缺值的 `UsageError`
- [ ] `go test ./internal/opt` 全绿
- [ ] 五个工具接入 `opt.Parse`

## echo — [规范](specs/echo.md)

- [x] 规范与测试就绪
- [ ] 前置选项解析（`-n`/`-e`/`-E`、组合、左到右覆盖）
- [ ] 参数拼接与结尾换行
- [ ] 转义表（`\n`、`\t`、`\0NNN`、`\xHH`、`\c` 等）
- [ ] 未知转义与字面量参数
- [ ] `go test ./internal/echo` 全绿
- [ ] `go run ./cmd/echo -e "a\tb"` 手动验证

## cat — [规范](specs/cat.md)

- [x] 规范与测试就绪
- [ ] 短选项解析与 `--`
- [ ] stdin、`-`、多文件拼接
- [ ] `-n` / `-b` 编号（宽度 6 + TAB，跨文件连续）
- [ ] `-s` 压缩空行（全局状态）
- [ ] `-E` / `-T` 显示变换
- [ ] 文件错误后继续并返回 1
- [ ] `go test ./internal/cat` 全绿
- [ ] `go run ./cmd/cat -n README.md` 手动验证

## tee — [规范](specs/tee.md)

- [x] 规范与测试就绪
- [ ] `-a` 解析与 `--`
- [ ] 一次读取、stdout + 多文件写出
- [ ] 追加/截断/自动创建
- [ ] 单文件失败后继续其余目标并返回 1
- [ ] 二进制透传
- [ ] `go test ./internal/tee` 全绿
- [ ] `"hi" | go run ./cmd/tee tee-out.txt` 手动验证

## head — [规范](specs/head.md)

- [x] 规范与测试就绪
- [ ] `-n` / `-c` 解析、附加值、最后生效
- [ ] `-q` / `-v` 与文件头规则
- [ ] 按行/按字节截断、末尾无换行
- [ ] 多文件、stdin 操作数、错误后继续
- [ ] `go test ./internal/head` 全绿
- [ ] `go run ./cmd/head -n 3 README.md` 手动验证

## tail — [规范](specs/tail.md)

- [x] 规范与测试就绪
- [ ] `-n` / `-n +N` / `-c` 解析与最后生效
- [ ] 尾部行/字节选择、末尾无换行
- [ ] `-q` / `-v` 与文件头规则
- [ ] 多文件、stdin 操作数、错误后继续
- [ ] `go test ./internal/tail` 全绿
- [ ] `go run ./cmd/tail -n 3 README.md` 手动验证

## wc — [规范](specs/wc.md)

- [x] 规范与测试就绪
- [ ] `-l` / `-w` / `-c` / `-m` / `-L` 解析
- [ ] 行/词/字节/rune/最长行计数
- [ ] 固定列顺序与单空格格式
- [ ] 多文件 `total`、stdin 与 `-`
- [ ] 错误后继续并汇总可读文件
- [ ] `go test ./internal/wc` 全绿
- [ ] `go run ./cmd/wc README.md` 手动验证

## 收尾检查

- [ ] `gofmt -l .` 无输出
- [ ] `go vet ./...` 通过
- [ ] `go build ./...` 通过
- [ ] `go test ./...` 全部通过
- [ ] `pwsh -NoProfile -File scripts/check.ps1` 通过
- [ ] README、spec、测试三者无行为差异

## 架构决策

- [x] 已决定并落地 `internal/opt` 脚手架（类型、`UsageError`、25 个测试、规范、`Parse`
  存根）。设计说明见 [architecture.md §5.4](architecture.md)。
- [ ] 实现 `Parse` 并让五个工具接入；接入后工具验收测试保持不变。

## 进阶练习（不在验收测试内）

完成后可选做，并自行补测试：

- echo：`--help`/`--version`、`\u`/`\U`、更多八进制形式。
- cat：`-v`、`-A`、`-e`、`-t`、`-u`；严格流式处理大文件。
- tee：`-i`、`-p`、`--output-error`。
- head：负数计数（`-n -N`、`-c -N`）、长选项。
- tail：`-f` / `-F`、`--pid`、`-c +N`、环形缓冲实现。
- wc：GNU 风格列宽对齐、`--files0-from`。
