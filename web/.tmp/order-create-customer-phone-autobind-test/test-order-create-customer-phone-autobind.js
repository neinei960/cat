"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const pagePath = node_path_1.default.resolve(__dirname, '../../src/pages/order/create.vue');
const apiPath = node_path_1.default.resolve(__dirname, '../../src/api/order.ts');
const pageSource = node_fs_1.default.readFileSync(pagePath, 'utf8');
const apiSource = node_fs_1.default.readFileSync(apiPath, 'utf8');
(0, strict_1.default)(apiSource.includes('customer_phone?: string'), 'order create API payload should accept customer_phone for guest retail orders');
(0, strict_1.default)(pageSource.includes('function normalizePhoneInput'), 'order create should normalize typed phone input before matching');
(0, strict_1.default)(pageSource.includes('async function resolveCustomerByPhoneBeforeSubmit'), 'order create should resolve an existing customer by phone before submit');
(0, strict_1.default)(pageSource.includes('await resolveCustomerByPhoneBeforeSubmit(typedCustomerPhone)'), 'order create submit should await phone auto-binding before building payload');
(0, strict_1.default)(pageSource.includes('customer_phone: customerId ? undefined : typedCustomerPhone || undefined'), 'order create payload should send customer_phone only when no customer is selected');
console.log('order create customer phone auto-bind checks passed');
