## ADDED Requirements

### Requirement: 保存功能套餐 MUST NOT 新建商品编码

Admin 更新功能 SKU 时，商品编码去空白后为空，或 `feature_product` 中不存在该编码，系统 MUST 拒绝且 MUST NOT 插入新行。编码已存在时 MUST 只更新该行的价格、天数、Apple 商品 ID、上下架及其他非主键字段，MUST NOT 修改 `product_code`。

#### Scenario: 空编码保存

- **WHEN** 管理员提交的商品编码为空
- **THEN** 系统 MUST 拒绝，且 `feature_product` 行数 MUST 不变

#### Scenario: 已有编码改价

- **WHEN** 管理员提交已存在的商品编码，并把价格改为新的分值
- **THEN** 该编码的价格 MUST 更新，且 MUST NOT 出现新的商品编码

### Requirement: 打开功能时 MUST 填入已有套餐

开通功能管理选中某个功能时，售卖套餐表单 MUST 填入该功能已有套餐的商品编码。若存在上架且编码不以 `fp_` 开头的行，MUST 优先填入该行；否则填入上架行；再否则填入列表中的第一条。该功能没有任何套餐时，编码 MUST 保持为空，且页面 MUST NOT 生成新编码。页面 MUST NOT 再提供清空编码以新建套餐的操作。

#### Scenario: 打开已有种子套餐的功能

- **WHEN** 管理员打开一个已有非 `fp_` 上架套餐的功能
- **THEN** 表单商品编码 MUST 为该套餐编码

#### Scenario: 没有新建入口

- **WHEN** 管理员查看售卖套餐区域
- **THEN** 页面 MUST NOT 显示「新建套餐」
