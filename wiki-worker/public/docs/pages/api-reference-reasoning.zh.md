# 推理與模式符號

逐請求設定推理強度與服務層級的常數與函式。

## 推理

| 符號 | 簽名 |
|---|---|
| `Reasoning` 常數 | `ReasoningNone`、`ReasoningLow`、`ReasoningMedium`、`ReasoningHigh`、`ReasoningXHigh`、`ReasoningMax`，預設 `ReasoningDefault = ReasoningMedium` |
| `(Reasoning).String` | `func() string` |
| `ParseReasoning` | `func(s string) (Reasoning, bool)`，別名 `minimal` / `extra` / `ultra` |
| `ClampReasoning` | `func(r, lo, hi Reasoning, provider, model string) Reasoning` |
| `OpenAIEffortRange` | `func(model string) (low, high Reasoning)` |

## 模式

| 符號 | 簽名 |
|---|---|
| `Mode` 常數 | `ModeDefault`、`ModeFast` |
| `(Mode).String` | `func() string` |
| `ParseMode` | `func(s string) (Mode, bool)` |
| `SupportFast` | `func(providerName, model string) bool` |
| `WarnFastDowngrade` | `func(providerName, model, tier string)` |

## 相關頁面

- [推理等級](/zh/core-concepts-reasoning)
- [fast 模式](/zh/core-concepts-fast-mode)
