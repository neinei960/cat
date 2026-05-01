"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = __importDefault(require("fs"));
const path_1 = __importDefault(require("path"));
const filePath = path_1.default.resolve(process.cwd(), 'src/pages/order/batch-create.vue');
const source = fs_1.default.readFileSync(filePath, 'utf8');
function assertContains(pattern, message) {
    if (!pattern.test(source)) {
        throw new Error(message);
    }
}
assertContains(/<view\s+v-if="!productOrServiceModalOpen"\s+class="submit-bar">/, 'submit bar must be hidden while the add-service or add-product modal is open');
assertContains(/const\s+productOrServiceModalOpen\s*=\s*computed\(\(\)\s*=>\s*showAddService\.value\s*\|\|\s*showAddProduct\.value\)/, 'modal-open computed state must cover both service and product modals');
assertContains(/async function submitBatch\(\) \{\s*if \(productOrServiceModalOpen\.value\) return/, 'submitBatch must ignore clicks while a modal is open');
assertContains(/@click\.stop="setProductCategory\(0\)"/, 'the all-products category tab click must not bubble outside the modal');
assertContains(/@click\.stop="setProductCategory\(cat\.ID\)"/, 'product category tab clicks must not bubble outside the modal');
