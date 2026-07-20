"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const pagePath = node_path_1.default.resolve(__dirname, '../../src/pages/order/create.vue');
const source = node_fs_1.default.readFileSync(pagePath, 'utf8');
(0, strict_1.default)(source.includes('const PRODUCT_COLLAPSED_LIMIT = 5'), 'order create should define a five-product collapsed limit');
(0, strict_1.default)(source.includes('const productListExpanded = ref(false)'), 'order create should track whether the product list is expanded');
(0, strict_1.default)(source.includes('const visibleProductCards = computed'), 'order create should derive the displayed product cards separately from all filtered products');
(0, strict_1.default)(source.includes('v-for="product in visibleProductCards"'), 'order create should render the collapsed product card list instead of all filtered products');
(0, strict_1.default)(source.includes('展开全部') && source.includes('收起商品'), 'order create should expose expand and collapse controls for long product lists');
console.log('order create product collapse checks passed');
