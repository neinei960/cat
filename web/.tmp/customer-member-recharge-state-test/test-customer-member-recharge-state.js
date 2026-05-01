"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const filePath = node_path_1.default.resolve(__dirname, '../../src/pages/customer/detail.vue');
const source = node_fs_1.default.readFileSync(filePath, 'utf8');
(0, strict_1.default)(source.includes('const openCardAmount = ref(\'\')'), 'customer detail should use a dedicated amount state for opening a member card');
(0, strict_1.default)(source.includes('function openRechargeModal()'), 'customer detail should expose an explicit recharge modal opener');
(0, strict_1.default)(source.includes('function closeRechargeModal()'), 'customer detail should expose an explicit recharge modal closer');
(0, strict_1.default)(source.includes('openCardAmount.value') && !source.includes('parseFloat(rechargeAmount.value) < tpl.min_recharge'), 'member card threshold validation should only read the open-card amount state');
(0, strict_1.default)(source.includes('rechargeAmount.value = \'\''), 'recharge modal flow should clear stale recharge amount state');
console.log('customer member recharge state regression test passed');
