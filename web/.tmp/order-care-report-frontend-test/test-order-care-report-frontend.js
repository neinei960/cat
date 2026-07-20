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
    const apiSource = node_fs_1.default.readFileSync(node_path_1.default.join(root, 'src/api/order-care-report.ts'), 'utf8');
    assert(modalSource.includes("from '@/api/order-care-report'"), 'modal should import backend order care report api');
    assert(modalSource.includes('createOrderCareReport('), 'modal should call backend order care report api');
    assert(!modalSource.includes("from '@/api/pet-bath-report'"), 'modal should not import the pet bath report api');
    assert(!modalSource.includes('createPetBathReport('), 'modal should let the backend persist the pet bath report');
    assert(utilSource.includes('buildOrderCareReportPayload'), 'frontend should map the editor draft to the backend payload');
    assert(apiSource.includes("method: 'POST'"), 'order care report api should use POST');
    assert(apiSource.includes('/care-report'), 'order care report api should target the care report endpoint');
    assert(!stageSource.includes("from 'html2canvas'"), 'preview stage should not depend on browser screenshot rendering');
    assert(!stageSource.includes('exportPngBlob'), 'preview stage should not expose a browser screenshot exporter');
    assert(stageSource.includes("standard: { x: 732, y: 833 }"), 'body preview checkbox should use the real template center');
    assert(stageSource.includes("normal: { x: 406, y: 929 }"), 'skin preview checkbox should use the real template center');
    assert(stageSource.includes("dry: { x: 732, y: 1025 }"), 'hair preview checkbox should use the real template center');
    assert(stageSource.includes("cleaned: { x: 406, y: 1217 }"), 'eyes preview checkbox should use the real template center');
    assert(stageSource.includes("black_earwax: { x: 1058, y: 1313 }"), 'ears preview checkbox should use the real template center');
    assert(stageSource.includes("tartar: { x: 732, y: 1410 }"), 'oral preview checkbox should use the real template center');
    assert(stageSource.includes("normal: { x: 406, y: 1553 }"), 'anus preview checkbox should use the real template center');
    assert(stageSource.includes('care-report-stage-label-override'), 'preview should replace the old last-care label');
    assert(stageSource.includes('护理内容') && stageSource.includes('Content of care'), 'preview should show the real care-content labels');
    assert(stageSource.includes('formatDisplayDate'), 'preview should use the real dotted date format');
    assert(stageSource.includes('fontWeight: 700'), 'preview primary fields should use bold text');
    assert(stageSource.includes('left: x - 11') && stageSource.includes('top: y - 15'), 'preview checkmark should use the authentic larger bounds');
    assert(stageSource.includes('border-right: 5px solid #111111'), 'preview checkmark should use a bold stroke');
    assert(stageSource.includes("whiteSpace: 'normal'") && stageSource.includes("lineHeight: '24px'"), 'preview notes should support two centered lines');
    assert(stageSource.includes("createCenteredField('pet_name', props.draft.petName, 279, 196"), 'pet name preview should sit above its underline');
    assert(stageSource.includes("createCenteredField('care_date', formatDisplayDate(props.draft.careDate), 279, 686"), 'care date preview should sit above its underline');
    assert(stageSource.includes("createCenteredField('next_care_date', formatDisplayDate(props.draft.nextCareDate), 905, 690, 229, 48, 34, 400)"), 'next care date preview should use the lighter real-template weight');
    assert(stageSource.includes("{ key: 'skin', x: 648, y: 941"), 'skin preview note should stay centered around its underline');
    assert(stageSource.includes("{ key: 'ears', x: 648, y: 1325"), 'ears preview note should stay centered around its underline');
    assert(stageSource.includes("{ key: 'oral', x: 494, y: 1469"), 'oral preview note should stay centered around its underline');
    assert(stageSource.includes("{ key: 'anus', x: 494, y: 1566"), 'anus preview note should stay centered around its underline');
    assert(stageSource.includes("height: NOTE_LINE_HEIGHT,\n        display: 'flex',\n        alignItems: 'center',\n        justifyContent: 'center',\n        textAlign: 'center'"), 'section note previews should be centered on one or two lines');
}
main();
function assert(condition, label) {
    if (!condition) {
        throw new Error(label);
    }
}
