"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const node_fs_1 = require("node:fs");
const node_path_1 = require("node:path");
function assert(condition, message) {
    if (!condition) {
        throw new Error(message);
    }
}
const pagePath = (0, node_path_1.resolve)(process.cwd(), 'src/pages/order/batch-create.vue');
const source = (0, node_fs_1.readFileSync)(pagePath, 'utf8');
assert(source.includes("import { getStaffList } from '@/api/staff'"), 'batch order edit should load staff options');
assert(source.includes('const selectedStaffIdx = ref(0)'), 'batch order edit should track selected staff index');
assert(source.includes('const selectedStaff = computed'), 'batch order edit should expose selected staff');
assert(source.includes('<picker :range="staffNames"'), 'batch order edit should render a staff picker');
assert(source.includes('staff_id: selectedStaff.value?.ID'), 'batch order save should submit the currently selected staff');
assert(!source.includes('staff_id: existingOrder.value?.staff_id || appt.value?.staff_id'), 'batch order save must not keep the old staff_id when user changes staff');
console.log('order batch staff edit static checks passed');
