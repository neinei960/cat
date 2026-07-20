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
assertContains('v-if="petCareNoteItems.length"', 'appointment detail should render pet care notes when the pet profile has them');
assertContains('pet?.care_notes', 'appointment detail should read care notes from appointment pet profile data');
console.log('appointment detail care notes visibility checks passed');
