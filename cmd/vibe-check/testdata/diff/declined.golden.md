# 🏗️ Structural Entropy Diff

**Verdict**: REQUEST_CHANGES | **Entropy direction**: degrading

## ⚠️ Verdict Reasons

- new-cycle: ex/b ex/c
- instability: module "ex/a" delta 0.3000 >= threshold 0.1500
- distance: module "ex/a" delta 0.2500 >= threshold 0.2000
- lcom: module "ex/a" delta 2 >= threshold 2

## 📊 Per-module Deltas (PR minus base)

| Package | Instability | Abstractness | Distance | LCOM4 | Ca | Ce |
|---------|-------------|--------------|----------|-------|----|----|
| `ex/a` | 0.3000 | 0.0000 | 0.2500 | 2 | 0 | 0 |
| `ex/b` | 0.0000 | 0.0000 | 0.0000 | 0 | 0 | 0 |
| `ex/c` | 0.0000 | 0.0000 | 0.0000 | 0 | 0 | 0 |

## 🔄 New Cycles

🔴 **1 cycle detected:**

```
ex/b → ex/c → ex/b
```

## 🔄 Resolved Cycles

None.
