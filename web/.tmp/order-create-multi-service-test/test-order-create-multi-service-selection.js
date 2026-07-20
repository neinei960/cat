"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = require("fs");
const path_1 = require("path");
const assert_1 = __importDefault(require("assert"));
const source = (0, fs_1.readFileSync)((0, path_1.resolve)(process.cwd(), 'src/pages/order/create.vue'), 'utf8');
(0, assert_1.default)(source.includes('const serviceItems = ref<ServiceCartItem[]>([])'), 'order create should keep selected services in an array so one order can include multiple services');
(0, assert_1.default)(source.includes('function dedupeServicePriceRules'), 'order create should dedupe service price rules before rendering specs');
(0, assert_1.default)(source.includes('for (const item of serviceItems.value)'), 'order create submit should send every selected service as an order item');
(0, assert_1.default)(!source.includes('const serviceSubtotal = computed(() => selectedServiceId.value > 0 ? roundCurrency(servicePrice.value) : 0)'), 'order create service subtotal should not depend on a single selected service id');
console.log('order create multi-service selection checks passed');
