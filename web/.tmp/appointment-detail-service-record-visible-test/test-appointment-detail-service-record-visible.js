"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = require("fs");
const path_1 = require("path");
const source = (0, fs_1.readFileSync)((0, path_1.resolve)(process.cwd(), 'src/pages/appointment/detail.vue'), 'utf8');
function assertContains(text, message) {
    if (!source.includes(text)) {
        throw new Error(message);
    }
}
assertContains('v-if="shouldShowServiceRecords"', 'service records card should be visible when records exist, not only by appointment status');
assertContains('serviceRecords.value.length > 0', 'service records visibility should include existing service records');
assertContains('[1, 2, 3, 7].includes(Number(appt.value?.status || 0))', 'service records add button should be visible for confirmed, in-service, settlement, and billed appointments only');
assertContains('canAddServiceRecord', 'add service record button visibility should be captured in a named computed guard');
assertContains('v-if="canAddServiceRecord"', 'add service record button should be visible for confirmed and later active appointment states');
console.log('appointment detail service record visibility checks passed');
