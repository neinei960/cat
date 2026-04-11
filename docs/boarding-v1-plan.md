# 寄养功能 V1 实施方案

## Summary

在现有 `customer / pet / order / staff` 体系上新增一条"寄养业务线"，实现柜子管理、寄养优惠策略、寄养开单、在住看板和寄养订单详情。V1 目标是让门店可以维护柜子资源，给老客或新客的猫开寄养单，按入住日期占柜，选择并叠加优惠策略，最后复用现有 `orders` 完成收银。

本方案锁定以下业务决定：

- 寄养复用现有订单体系，不单独做支付体系。
- 首版支持同柜多猫，但必须由柜型和容量控制，且同柜多猫必须属于同一客户。
- 离店时才按实际住宿天数结算收款，不预收全款。
- 优惠只有两种：住N免M、节假日加收。
- 节假日不复用现有预约日历展示数据，改为店长在后台维护店铺节假日配置，作为寄养计价唯一数据源。

## Key Changes

### 后端模型与业务

- 新增 `boarding_cabinets`
  - 字段：`id, shop_id, code, cabinet_type, capacity, base_price, status, remark, created_at, updated_at, deleted_at`
  - `cabinet_type`：柜子类型名称，如 "A柜"、"B柜"
  - `status` 固定为：`enabled / cleaning / disabled`
  - 柜子在看板中的展示状态 `idle / reserved / occupied` 由当前时间和有效寄养单实时推导，不单独落库
- 新增 `boarding_holidays`
  - 字段：`id, shop_id, holiday_date, name, created_at, updated_at, deleted_at`
  - 作为寄养节假日加收的唯一数据源，由店长维护
- 新增 `boarding_discount_policies`
  - 字段：`id, shop_id, name, policy_type, rule_json, valid_from, valid_to, priority, stackable, status, remark, created_at, updated_at, deleted_at`
  - V1 支持策略类型：
    - `stay_n_free_m`：住满 N 天免 M 天，免费天数从非节假日天数中扣减
    - `holiday_surcharge`：节假日每晚加收固定金额
  - `rule_json` 示例：
    - `stay_n_free_m`：`{"stay": 5, "free": 1}`
    - `holiday_surcharge`：`{"surcharge": 30}`（每晚加收 30 元）
  - 节假日日期由 `boarding_holidays` 提供，不在策略中单独配置
- 新增 `boarding_orders`
  - 字段：`id, shop_id, order_id, customer_id, staff_id, cabinet_id, check_in_at, check_out_at, actual_check_out_at, nights, base_amount, holiday_surcharge_amount, discount_amount, pay_amount, status, remark, policy_snapshot_json, price_snapshot_json, created_at, updated_at, deleted_at`
  - `check_out_at`：预定离店时间
  - `actual_check_out_at`：实际离店时间，离店结算时填入
  - `status` 固定为：`pending_checkin / checked_in / checked_out / cancelled`
- 新增 `boarding_order_pets`
  - 字段：`id, boarding_order_id, pet_id, pet_name_snapshot, remark`
  - 用来支持同柜多猫，且保留开单时快照
- 新增 `boarding_order_logs`
  - 记录入住、续住、换柜、离店等动作
- 复用现有 `orders` 和 `order_items`
  - `orders` 负责订单号、支付状态、统一订单列表展示
  - `order_items` 新增寄养相关明细类型：`4=寄养住宿`、`5=节假日加收`、`6=住N免M抵扣`
  - 订单状态规则固定：
    - 创建寄养单时同步创建一张现有 `order`，金额使用当前预估金额，`pay_status=0`、`status=0`
    - 在客户未付款前，续住、换柜、离店重算都直接覆盖这张未支付订单的 `total_amount / discount_amount / pay_amount / order_items`
    - 办理离店时，以 `actual_check_out_at` 重算最终金额并同步订单明细，然后在该订单上完成收款
    - 未入住前取消寄养单时，同步将现有订单置为取消
    - 一旦该寄养订单对应的现有订单完成支付，金额和明细锁定，不再允许继续续住或改柜，只能走退款后重开

### 价格与优惠规则

- 寄养价格按"天数"计算（住宿天数 = 离店日期 - 入住日期）
- 价格计算流程：
  1. 按日期拆分普通日与节假日，每天基础价 = 柜子 `base_price`
  2. 节假日天数 × `surcharge` 金额 = 节假日加收总额
  3. 应用 `stay_n_free_m`：总天数满 N 天则免 M 天，免费天数从非节假日天数中扣减（扣减的是 `base_price`）
  4. 最终金额 = 基础住宿总额 + 节假日加收总额 - 免费天数抵扣金额
- 叠加规则：
  - `holiday_surcharge` 和 `stay_n_free_m` 可以叠加
  - 同类型策略同一笔订单只取一条，按 `priority` 最大者生效
- 每次开单必须落价格快照，后续修改策略不影响历史订单
- **离店结算**：离店时填入 `actual_check_out_at`，按实际天数重新跑价格计算流程，生成最终金额后收款
- **订单同步**：每次价格重算后，必须同步覆盖未支付 `orders` 的金额字段和 `order_items` 明细，保证寄养单和订单单据金额完全一致

### 柜子占用与同柜多猫

- 柜子可用性按时间区间判断，不按当前状态静态判断
- 同一柜子在同一时间段只能存在一张有效 `boarding_order`
- 同柜多猫通过 `boarding_order_pets` 实现，同一订单内多只猫必须属于同一客户
- `capacity` 表示该柜最大猫数；开单时选中的猫数量不得超过 `capacity`
- 换柜是修改 `boarding_order.cabinet_id` 并写日志，不拆单
- 续住直接修改 `check_out_at` 日期，需重新校验当前柜子在续住时段仍可用
- 柜子看板展示状态规则固定：
  - `disabled / cleaning` 直接读取柜子表 `status`
  - `occupied`：当前时间处于某张 `checked_in` 寄养单的入住区间内
  - `reserved`：当前时间之前未入住，但未来存在 `pending_checkin` 且未取消的寄养单
  - `idle`：除以上情况外默认空闲

### 前端页面

- 新增页面
  - `寄养看板`：按柜子类型分组展示，每组显示该类型下所有柜子的当前状态（在住/待入住/空闲/清洁中），在住柜子显示猫咪信息和预计离店时间
  - `寄养开单`：选客户、选猫、选柜、选策略、预览价格、生成订单
  - `寄养订单详情`：价格明细、猫咪列表、柜子、状态流转、操作记录
  - `柜子管理`：增删改柜子，按类型分组展示
  - `节假日管理`：店长维护哪些日期算节假日
  - `优惠策略管理`：店长配置规则
- 导航建议
  - 底部现有"订单"保持不变
  - 在"更多"或"工作台"新增 `寄养` 入口
- 开单页面流程固定
  1. 搜索老客或创建新客
  2. 选择猫咪，老客从已有猫中勾选，新客支持现场建猫
  3. 选择入住时间和预计离店时间，自动计算天数
  4. 拉取可用柜子列表并选择柜子
  5. 自动列出命中的优惠策略，允许勾选可叠加策略
  6. 输入备注
  7. 生成寄养业务单和现有订单

## Public APIs / Interfaces

- `GET /b/boarding/cabinets`
- `POST /b/boarding/cabinets`
- `PUT /b/boarding/cabinets/:id`
- `GET /b/boarding/cabinets/availability`
  - 入参：`check_in_at, check_out_at, pet_count`
  - 返回满足容量与时间区间的柜子，按类型分组
- `GET /b/boarding/holidays`
- `POST /b/boarding/holidays`
- `DELETE /b/boarding/holidays/:id`
- `GET /b/boarding/policies`
- `POST /b/boarding/policies`
- `PUT /b/boarding/policies/:id`
- `POST /b/boarding/orders/price-preview`
  - 入参：`customer_id, pet_ids[], cabinet_id, check_in_at, check_out_at, policy_ids[]`
  - 返回：基础住宿明细、节假日加收明细、免费天数抵扣明细、应付金额、命中策略说明
- `POST /b/boarding/orders`
  - 入参：`customer_id, pet_ids[], cabinet_id, check_in_at, check_out_at, policy_ids[], remark`
  - 创建 `boarding_order + boarding_order_pets + orders + order_items`
- `GET /b/boarding/orders`
- `GET /b/boarding/orders/:id`
- `GET /b/boarding/dashboard`
  - 返回所有柜子按类型分组的当前状态及在住信息
- `PUT /b/boarding/orders/:id/check-in`
- `PUT /b/boarding/orders/:id/check-out`
  - 入参：`actual_check_out_at`
  - 按实际天数重算价格，更新订单金额，完成结算
- `PUT /b/boarding/orders/:id/extend`
  - 入参：新的 `check_out_at`
  - 校验柜子可用性，更新预计离店日期
- `PUT /b/boarding/orders/:id/change-cabinet`
  - 入参：新的 `cabinet_id`
  - 校验新柜子可用性，更新柜子并写日志
- `PUT /b/boarding/orders/:id/cancel`
- 权限
  - `admin`：柜子管理、节假日管理、策略管理、改价
  - `manager`：寄养开单、入住、离店、续住、换柜
  - `staff`：查看看板、查看详情，不可改策略

## Test Plan

- 柜子管理
  - 新增、编辑、停用柜子后列表正确展示
  - 停用柜子不能再出现在可用柜子列表
  - 按类型分组展示正确
- 节假日管理
  - 店长新增、删除节假日后，价格预览和离店重算立即生效
  - 非节假日日期不会触发 `holiday_surcharge`
- 开单
  - 老客单猫开单成功并生成现有订单
  - 新客现场建档后开单成功
  - 同柜两猫开单成功，超出 `capacity` 时拦截
  - 某柜在时间区间已被占用时不可再选
  - 预览接口能校验多只猫属于同一客户
- 优惠
  - 节假日加收生效，金额正确
  - 住N免M 单独生效，免费天数从非节假日扣减
  - 两种策略叠加后金额正确
  - 同类型多策略命中时只取 `priority` 最大者
  - 历史单在策略变更后金额不回算
- 订单联动
  - 创建寄养单时自动生成一张未支付订单
  - 续住、换柜、提前离店后未支付订单金额与明细被正确覆盖
  - 取消未入住寄养单时现有订单同步取消
  - 已支付寄养单不可再续住或换柜
- 离店结算
  - 按预计日期离店，金额与开单时一致
  - 提前离店，按实际天数重算，优惠不满足条件时不再命中
  - 延后离店（先续住再离店），金额正确
- 生命周期
  - 待入住单可办理入住
  - 在住单可续住、换柜、离店
  - 待入住单可取消
- 回归
  - 现有洗护预约开单、订单列表、支付、退款不受影响

## Assumptions

- V1 不收押金，离店时一次性结算收款
- V1 不做"一个柜子拆成多个独立床位"的复杂资源模型，只做单柜容量
- V1 不做会员卡自动折上折；会员寄养价如果要支持，作为一条寄养优惠策略配置
- V1 的节假日由店长在寄养模块内维护，不复用预约日历展示数据
- V1 不支持附加项（加餐、洗澡等），后续版本再考虑
