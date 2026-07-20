"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const root = process.cwd();
const pageSource = node_fs_1.default.readFileSync(node_path_1.default.join(root, 'src/pages/order/detail.vue'), 'utf8');
const apiSource = node_fs_1.default.readFileSync(node_path_1.default.join(root, 'src/api/order.ts'), 'utf8');
(0, strict_1.default)(apiSource.includes('updateOrderCustomerPet'), 'order API should expose a dedicated customer/pet update call');
(0, strict_1.default)(apiSource.includes('/customer-pet'), 'customer/pet update should call the dedicated order endpoint');
(0, strict_1.default)(pageSource.includes('showCustomerPetModal'), 'order detail should track customer/pet edit modal state');
(0, strict_1.default)(pageSource.includes('修改客户/猫咪'), 'order detail should provide a visible customer/pet edit action');
(0, strict_1.default)(pageSource.includes('saveCustomerPet'), 'order detail should save customer/pet edits without opening the full order editor');
(0, strict_1.default)(pageSource.includes('updateOrderCustomerPet(order.value.ID'), 'order detail should persist customer/pet edits through the dedicated API');
console.log('order detail customer/pet edit checks passed');
