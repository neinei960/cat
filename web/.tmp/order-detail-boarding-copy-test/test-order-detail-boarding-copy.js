"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const filePath = node_path_1.default.resolve(__dirname, '../../src/pages/order/detail.vue');
const source = node_fs_1.default.readFileSync(filePath, 'utf8');
(0, strict_1.default)(source.includes("if (order.value?.order_kind !== 'boarding') {"), 'boarding header pet list should skip petGroups-derived room names');
(0, strict_1.default)(source.includes("pet_name: getReceiptGroupName(group.pet_name, order.value?.order_kind)"), 'boarding detail groups should reuse normalized group names');
(0, strict_1.default)(source.includes("name: getReceiptItemDisplayName(item.name, group.pet_name === '零售商品', retailNamePrefixes.value, order.value?.order_kind)"), 'boarding detail items should reuse normalized item names with order kind context');
console.log('order-detail boarding copy tests passed');
