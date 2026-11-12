# opt 规范（共享参数解析器）

## 1. 目标与状态

`internal/opt` 是 cat/head/tail/wc/tee 共用的**短选项语法扫描器**。它只负责把命令行切成“选项出现序列 +
操作数”，不关心数字是否合法、哪个选项覆盖哪个、文件怎么读。

当前状态：

- 已提供类型、`UsageError`、`Parse` 存根、25 个验收测试和本文档。
- `Parse` 仍返回 `opt: not implemented`，所以 `go test ./internal/opt` 是红的。
- 五个工具尚未接入；echo 不使用本包。

## 2. API

```go
package opt

type Option struct {
	Short      byte
	TakesValue bool
}

type Occurrence struct {
	Short byte
	Value string // 无值选项时为空
}

type Result struct {
	Operands []string     // 保持命令行顺序
	Options  []Occurrence // 保持命令行顺序，供工具实现 last-wins
}

type UsageError struct {
	Prog string
	Msg  string
}

func (e *UsageError) Error() string

func Parse(prog string, options []Option, args []string) (Result, error)
```

- `prog` 只用于错误消息前缀。
- `options` 声明本工具接受的短选项；可以为空。
- `args` 是程序名之后的参数。
- 返回值不包含退出码；语法错误时工具打印错误并返回 `cli.ExitUsage`。

## 3. 选项声明

- `Short` 是一个字节，实践中为 ASCII 字母。
- `TakesValue == false`：无值选项，可和其他无值选项组合。
- `TakesValue == true`：带值选项，可接受附加值或下一个参数。
- 不要把 `-` 声明为 `Short`：单独的 `-` 按操作数处理。
- 同一 `Short` 不应重复声明；行为未定义且测试不覆盖。
- 不支持长选项、`--name=value` 或 GNU 的 `getopt_long` 语义。

## 4. 扫描规则

从 `args` 左到右扫描，全部解析完才返回：

1. 参数等于 `--`：结束选项解析，其余全部按操作数处理。
2. 参数等于 `-`：操作数。
3. 参数以 `-` 开头且长度大于 1：按短选项簇扫描，逐个处理 `-` 后的字节。
4. 其他参数：操作数，保持原始顺序。

选项簇规则：

- 无值选项：记录一次 `Occurrence{Short: c}`，继续扫描下一个字符。
- 带值选项：
    - 若该字符后还有剩余字符，**剩余全部**作为值，例如 `-n5` → `n=5`、`-nq` → `n=q`。
    - 否则消耗下一个参数作为值，即使它是 `-`、`--` 或以 `-` 开头，例如 `-n -5` → `n=-5`、`-n --` → `n=--`。
    - 没有下一个参数 → 返回缺值错误。
- 簇中遇到未声明的字符 → 返回未知选项错误，消息使用**完整原始参数**（例如 `-az`）。

顺序与结构：

- `Result.Options` 按命令行出现顺序保存，因此工具能实现“最后一个 `-n` / `-c` 生效”。
- `Result.Operands` 按出现顺序保存；选项和操作数可混排。
- 同一个选项出现多次时全部保留，不在解析器内合并。

## 5. 错误契约

只产生两类语法错误，都返回 `*opt.UsageError`：

| 场景           | `Msg`                                      | 完整错误示例（`prog = "head"`）         |
|----------------|--------------------------------------------|-----------------------------------------|
| 未知选项       | `unrecognized option: <完整原始 token>`    | `head: unrecognized option: -az`        |
| 带值选项缺少值 | `option requires an argument: -<选项字符>` | `head: option requires an argument: -n` |

- `UsageError.Error()` 返回 `"<prog>: <msg>"`；`Prog` 为空时只返回 `Msg`。
- 出错时的 `Result` 内容未定义；工具不应使用它。
- 工具不应自行改写这两类消息，否则会破坏与 `internal/opt` 测试的一致性。

## 6. 示例

```go
opts := []opt.Option{
{Short: 'n', TakesValue: true},
{Short: 'c', TakesValue: true},
{Short: 'q'},
}

opt.Parse("head", opts, []string{"a.txt", "-n", "1", "b.txt"})
// Result{
//   Operands: []string{"a.txt", "b.txt"},
//   Options:  []opt.Occurrence{{Short: 'n', Value: "1"}},
// }, nil

opt.Parse("head", opts, []string{"-qn5"})
// Result{Options: []opt.Occurrence{{Short: 'q'}, {Short: 'n', Value: "5"}}}, nil

opt.Parse("head", opts, []string{"-n", "--"})
// Result{Options: []opt.Occurrence{{Short: 'n', Value: "--"}}}, nil

opt.Parse("head", opts, []string{"-n"})
// nil 结果 + *opt.UsageError{Prog: "head", Msg: "option requires an argument: -n"}
```

工具侧写法：

```go
parsed, err := opt.Parse("head", headOptions, args)
if err != nil {
fmt.Fprintln(stderr, err)
return cli.ExitUsage
}
```

## 7. 测试对照

测试文件：`internal/opt/opt_test.go`，测试函数 `TestParse`。

| 子测试组                                                                                                   | 验证点                               |
|------------------------------------------------------------------------------------------------------------|--------------------------------------|
| `empty arguments`、`operands only`、`double dash alone`                                                    | 空输入、纯操作数、`--`               |
| `single flag`、`flag cluster`、`repeated flags preserved`                                                  | 单个/组合/重复选项                   |
| `attached value`、`separate value`、`empty separate value`                                                 | 附加值、独立值、空字符串值           |
| `value may start with dash`、`value consumes double dash`                                                  | 值可以以 `-` 开头、`--` 作为值被消耗 |
| `value option at end of cluster`、`value consumes rest of cluster`                                         | 值选项位于簇尾/簇中                  |
| `option operand permutation`、`operands after value`、`cluster of flags then operand`                      | 混排与操作数顺序                     |
| `double dash stops options`、`single dash is operand`                                                      | `--` 终止与单独 `-`                  |
| `preserves option order`                                                                                   | 多次选项的出现顺序（供 last-wins）   |
| `unknown option`、`unknown character in cluster`、`long option is unknown`、`unknown option after operand` | 未知选项与完整 token 消息            |
| `missing value at end`、`missing value after cluster flag`                                                 | 缺值消息使用选项字符                 |

## 8. 实现注意

- 只按字节扫描；选项是 ASCII 字节，不需要 Unicode 处理。
- 不要修改传入的 `args` 切片；操作数可以直接引用其中的字符串。
- `Result.Options` 必须按出现顺序 append，不要按选项类型分组。
- 带值选项的“剩余部分作为值”包括空字符串以外的任意字节；独立值允许空字符串。
- 未知选项消息保留原始 token，便于定位 `-az` 这类簇。
- 缺少值时报告带值选项本身的 `-n`，而不是整个簇（如 `-qn` 报 `-n`）。
- 不要用标准库 `flag` 实现本接口；语义不匹配。

## 9. 非目标

- 不校验数字（`head -n abc` 的 `invalid number` 由工具负责）。
- 不实现 last-wins / 覆盖规则（由工具遍历 `Result.Options` 决定）。
- 不处理文件、stdin/stdout、错误输出或退出码。
- 不支持长选项、`=` 形式、选项缩写。
- 不处理 echo 的特殊前置选项语义。

## 10. 落地状态

- [x] 类型、`UsageError`、`Parse` 存根
- [x] 25 个验收测试
- [x] 本规范
- [ ] 实现 `Parse`
- [ ] 五个工具接入 `opt.Parse`
- [ ] 删除 `Parse` 的 TODO 与 architecture 中的“待实现”标记
- [ ] `go test ./internal/opt` 全绿

## 11. 参考

- [Go `flag` 包](https://pkg.go.dev/flag)（说明为什么不能直接满足本语义）
- [GNU getopt 手册](https://www.gnu.org/software/libc/manual/html_node/Getopt.html)
- [architecture.md §5](../architecture.md)（选项扫描的公共规则）
