"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
function main() {
    const root = process.cwd();
    const modalSource = node_fs_1.default.readFileSync(node_path_1.default.join(root, 'src/components/order/OrderCareReportModal.vue'), 'utf8');
    const stageSource = node_fs_1.default.readFileSync(node_path_1.default.join(root, 'src/components/order/OrderCareReportStage.vue'), 'utf8');
    const utilSource = node_fs_1.default.readFileSync(node_path_1.default.join(root, 'src/utils/order-care-report.ts'), 'utf8');
    assert(!modalSource.includes("from '@/api/order-care-report'"), 'modal should not import backend order care report api');
    assert(!modalSource.includes('createOrderCareReport('), 'modal should not call backend order care report api');
    assert(modalSource.includes("from '@/api/pet-bath-report'"), 'modal should persist via pet bath report api');
    assert(modalSource.includes('createPetBathReport('), 'modal should create pet bath report record');
    assert(!utilSource.includes('buildOrderCareReportPayload'), 'frontend rendering flow should not expose backend payload builder');
    assert(stageSource.includes("shellElement?.querySelector('.care-report-stage')"), 'stage export should capture the inner report stage instead of the scaled preview shell');
    assert(stageSource.includes('const exportScale = rect.width ? REPORT_WIDTH / rect.width : 1'), 'stage export should scale the preview capture back to template resolution');
    assert(stageSource.includes('scale: exportScale'), 'stage export should apply the computed export scale');
    assert(stageSource.includes('width: rect.width'), 'stage export should capture the visible preview width before scaling up');
}
main();
function assert(condition, label) {
    if (!condition) {
        throw new Error(label);
    }
}
