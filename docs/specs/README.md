# 规范索引

每个工具的行为以这里的文档为唯一准绳；测试是这些规范的执行版本。

## 工具

| 工具 | 规范 | 测试 | 练习重点 |
|---|---|---|---|
| echo | [echo.md](echo.md) | `internal/echo/echo_test.go` | 参数解析、转义、字节输出 |
| cat | [cat.md](cat.md) | `internal/cat/cat_test.go` | 流式 I/O、行号、显示变换 |
| tee | [tee.md](tee.md) | `internal/tee/tee_test.go` | 一次读取、多路写出、失败降级 |
| head | [head.md](head.md) | `internal/head/head_test.go` | 按行/字节截断、文件头 |
| tail | [tail.md](tail.md) | `internal/tail/tail_test.go` | 尾部选择、`+N`、缓冲 |
| wc | [wc.md](wc.md) | `internal/wc/wc_test.go` | 计数、Unicode、固定列格式 |

## 阅读顺序

1. 先看本文和 [架构约定](../architecture.md) 中的公共规则（退出码、错误格式、参数扫描）。
2. 再看目标工具的“用法”“选项”“行为细节”。
3. 实现时对照“测试对照”表逐项验证。
4. 全部通过后阅读“进阶练习”，自行扩展。

## 写新 spec 的模板

```markdown
# <tool> 规范

## 1. 目标
## 2. 用法
## 3. 选项
## 4. 行为细节
### 4.1 输入来源
### 4.2 参数解析
### 4.3 处理规则
### 4.4 多文件与错误
### 4.5 输出格式
## 5. 退出码
## 6. 错误信息
## 7. 示例
## 8. 测试对照
## 9. 实现注意
## 10. 进阶练习
## 11. 参考
```

## 维护规则

- 修改行为时，同一次改动中更新 spec、测试和 `docs/progress.md`。
- 测试名称出现在 spec 的“测试对照”表中，重命名测试时同步更新。
- 新增工具时，在 `docs/progress.md` 和本索引中登记。
- 共享包（如计划中的 `internal/opt`）的设计见 [architecture.md §5.4](../architecture.md)；落地时再补对应 spec。
- 公共规则（选项扫描、错误格式、stdin 名称等）只在 [architecture.md](../architecture.md) 中定义，spec 只写差异和细节。
